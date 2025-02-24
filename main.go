package main

import (
	"asteroid/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sethvargo/go-envconfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

var cfg types.Config
var dbconn *mongo.Client
var rmq *amqp.Connection

func publishSnapshot(cam types.Camera, batch_id primitive.ObjectID, wg *sync.WaitGroup) types.Snapshot {
	defer wg.Done()

	snap := fetchSnapshot(cam)
	if snap == (types.Snapshot{}) {
		return snap
	}
	snap.BatchID = batch_id

	result, err := dbconn.Database(cfg.MongoDB.DB).Collection("snapshots").InsertOne(context.TODO(), snap)
	failOnError(err, "Failed to insert snapshot")
	snap.ID = result.InsertedID.(primitive.ObjectID)

	ch, err := rmq.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_queue",
		true,
		false,
		false,
		false,
		amqp.Table{"x-max-length": cfg.RabbitMQServer.MaxQueueLength},
	)
	failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg, err := json.Marshal(struct {
		ID       primitive.ObjectID `json:"_id"`
		URL      string             `json:"url"`
		Location string             `json:"location"`
		BatchID  primitive.ObjectID `json:"batch_id"`
	}{snap.ID, snap.URL, cam.Location, snap.BatchID})
	failOnError(err, "Failed to marshal snapshot")

	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			Body:        msg,
		})

	failOnError(err, "Failed to publish")
	log.Printf(" [x] Sent %s\n", snap.ID)

	return snap
}

func fetchSnapshot(cam types.Camera) types.Snapshot {
	t := time.Time.Local(time.Now())
	ret := types.Snapshot{
		Filename:  fmt.Sprintf("%s_%s.jpg", cam.ID.Hex(), t.Format("2006-01-02T15:04:05")),
		Timestamp: t,
		URL:       fmt.Sprintf("%s/%s_%s.jpg", cfg.Snapshot.URLPrefix, cam.ID.Hex(), t.Format("2006-01-02T15:04:05")),
		CameraID:  cam.ID,
	}

	cmd := exec.Command("ffmpeg", "-i", cam.URL, "-r", "1", "-frames:v", "1", fmt.Sprintf("%s/%s", cfg.Snapshot.Directory, ret.Filename))
	if err := cmd.Run(); err != nil {
		// print command output
		log.Printf("Failed to take snapshot for camera %s: %s", cam.ID, err)
		output, _ := cmd.CombinedOutput()
		log.Printf("Command output: %s \nStdERR: %s", string(output), string(err.(*exec.ExitError).Stderr))
		return types.Snapshot{}
	}

	return ret
}

func createBatch(waitInterval float32) types.Batch {
	// round current time to a multiple of waitInterval
	t := time.Now()
	t = t.Add(time.Duration(-t.Nanosecond()))
	t = t.Add(time.Duration(-t.Second() % int(waitInterval)))
	batch := types.Batch{
		Timestamp: t,
	}

	result, err := dbconn.Database(cfg.MongoDB.DB).Collection("batches").InsertOne(context.TODO(), batch)
	failOnError(err, "Failed to insert batch")
	batch.ID = result.InsertedID.(primitive.ObjectID)

	return batch
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var wg sync.WaitGroup

func main() {
	// check if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	failOnError(err, "ffmpeg is not installed")

	// check if config file exists
	if _, err := os.Stat("config.yaml"); os.IsExist(err) {
		f, err := os.Open("config.yaml")
		failOnError(err, "Failed to open config file")
		defer f.Close()
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&cfg)
		failOnError(err, "Failed to decode config file")
	} else {
		failOnError(envconfig.Process(context.Background(), &cfg),
			"Failed to process environment variables")
	}

	if cfg.RabbitMQServer.URI != "" {
		rmq, err = amqp.Dial(cfg.RabbitMQServer.URI)
	} else {
		rmq, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
			cfg.RabbitMQServer.User, cfg.RabbitMQServer.Password, cfg.RabbitMQServer.Host, cfg.RabbitMQServer.Port))
	}
	failOnError(err, "Failed to connect to RabbitMQ")
	defer rmq.Close()

	dbconn, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.MongoDB.URI))
	failOnError(err, "Failed to connect to MongoDB")

	camerasColl := dbconn.Database(cfg.MongoDB.DB).Collection("cameras")

	wg := sync.WaitGroup{}
	for {
		batch := createBatch(cfg.Snapshot.WaitInterval)
		cur, err := camerasColl.Find(context.TODO(), bson.D{{}})
		failOnError(err, "Failed to fetch cameras")
		fmt.Println("Fetching snapshots for batch", batch.ID, " ", batch.Timestamp)
		for cur.Next(context.TODO()) {
			var cam types.Camera
			err := cur.Decode(&cam)
			failOnError(err, "Failed to decode camera")
			wg.Add(1)
			go publishSnapshot(cam, batch.ID, &wg)
		}
		wg.Wait()
		// set batch fetch_completed to true
		_, err = dbconn.Database(cfg.MongoDB.DB).Collection("batches").UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: batch.ID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "fetch_completed", Value: true}}}})
		if err != nil {
			log.Printf("Failed to update batch %s: %s", batch.ID, err)
		} else {
			fmt.Println("Batch: ", batch.ID, " fetch completed")
		}
		time.Sleep(time.Duration(cfg.Snapshot.WaitInterval) * time.Second)
	}
}
