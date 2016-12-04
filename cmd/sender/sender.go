// Sender implementation for this Kubernetes RabbitMQ POD auto-scaling demo
// Based on https://github.com/rabbitmq/rabbitmq-tutorials
package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {

	// Command line options
	brokerptr := flag.String("b", "rabbitmq", "SVC (Kubernetes) name where to find the RabbitMQ broker (can also be an IP)")
	portptr := flag.String("p", "5672", "Port where RabbitMQ broker listens on")
	//fibptr := flag.Int("n", 10, "Specify the fibonacci number to process")
	flag.Parse()

	// Dereference the pointer from flag
	broker := *brokerptr
	port := *portptr

	randch := make(chan int)

	conn, err := amqp.Dial("amqp://guest:guest@" + broker + ":" + port)
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Println("Connected.")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	log.Println("Opened a channel")
	defer ch.Close()

	err = ch.Qos(
		1,    // prefetch count
		0,    // prefetch size
		true, // global
	)
	if err != nil {
		log.Println("Failed to set QoS")
	}

	q, err := ch.QueueDeclare(
		"fib-work", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")
	log.Println("Got a queue")

	go func() {

		for {
			i := rand.Intn(30)
			randch <- i

		}
	}()

	for {
		body := strconv.Itoa(<-randch)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})

		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %q", body)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	}

}
