package main

import (
	"asteroid/types"
	"context"
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

func publishSnapshot(cam types.Camera, wg *sync.WaitGroup) types.Snapshot {
	defer wg.Done()

	snap := fetchSnapshot(cam)

	result, err := dbconn.Database(cfg.MongoDB.DB).Collection("snapshots").InsertOne(context.TODO(), snap)
	failOnError(err, "Failed to insert snapshot")
	snap.ID = result.InsertedID.(primitive.ObjectID)

	ch, err := rmq.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(snap.ID.Hex()),
		})

	failOnError(err, "Failed to publish")
	log.Printf(" [x] Sent %s\n", snap.ID)

	return snap
}

func fetchSnapshot(cam types.Camera) types.Snapshot {
	t := time.Now()
	ret := types.Snapshot{
		Filename:  fmt.Sprintf("%s_%s.jpg", cam.ID.Hex(), t.Format("2006-01-02T15:04:05")),
		Timestamp: t,
		URL:       fmt.Sprintf("%s/%s_%s.jpg", cfg.Snapshot.URL_Prefix, cam.ID.Hex(), t.Format("2006-01-02T15:04:05")),
		CameraID:  cam.ID,
	}

	cmd := exec.Command("ffmpeg", "-i", cam.URL, "-r", "1", "-frames:v", "1", fmt.Sprintf("%s/%s", cfg.Snapshot.Directory, ret.Filename))
	if err := cmd.Run(); err != nil {
		// print command output
		log.Printf("Failed to take snapshot for camera %s: %s", cam.ID, err)
		log.Printf("Command output: %s", string(err.(*exec.ExitError).Stderr))
		return types.Snapshot{}
	}

	return ret
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// check if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	failOnError(err, "ffmpeg is not installed")

	f, err := os.Open("config.yaml")
	failOnError(err, "Failed to open config file")
	defer f.Close()

	failOnError(envconfig.Process(context.Background(), &cfg),
		"Failed to process environment variables")

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	failOnError(err, "Failed to decode config file")

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

	coll := dbconn.Database(cfg.MongoDB.DB).Collection("cameras")
	cur, err := coll.Find(context.TODO(), bson.D{{}})
	failOnError(err, "Failed to fetch cameras")

	var wg sync.WaitGroup
	for {
		for cur.Next(context.TODO()) {
			var cam types.Camera
			err := cur.Decode(&cam)
			failOnError(err, "Failed to decode camera")
			wg.Add(1)
			go publishSnapshot(cam, &wg)

		}
		time.After(30 * time.Second)
	}

	wg.Wait()
}
