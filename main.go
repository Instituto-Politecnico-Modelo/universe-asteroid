package main

import (
	"context"
	"go/types"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s:  %s", msg, err)
	}
}

func main() {
	f, err := os.Open("config.yaml")
	failOnError(err, "Failed to open config file")
	defer f.Close()

	var cfg types.Config

	conn, err := amqp.Dial("amqp://guest:guest@172.17.0.2:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
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

	body := "Hello World!"

	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	failOnError(err, "Failed to publish")
	log.Printf(" [x] Sent %s\n", body)
}
