// Sender implementation for this Kubernetes RabbitMQ POD auto-scaling demo
// Based on https://github.com/rabbitmq/rabbitmq-tutorials
package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func fib(number int) int {
	if number == 0 || number == 1 {
		return number
	}

	return fib(number-2) + fib(number-1)
}

func main() {

	// --- Command line options
	brokerptr := flag.String("b", "rabbitmq", "SVC (Kubernetes) name where to find the RabbitMQ broker")
	portptr := flag.String("p", "5672", "Port where RabbitMQ broker listens on")
	flag.Parse()

	// --- Dereference the pointer from flag
	broker := *brokerptr
	port := *portptr

	// --- RabbitMQ initialization
	// get a connection
	conn, err := amqp.Dial("amqp://guest:guest@" + broker + ":" + port)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// open channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// declare our queue
	q, err := ch.QueueDeclare(
		"fib-work", // name
		false,      // durable
		false,      // delete when usused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// set QoS on the channel (only fetch 1 message at a time)
	err = ch.Qos(
		1,    // prefetch count
		0,    // prefetch size
		true, // global
	)
	failOnError(err, "Failed to set QoS")

	// Set up consumer, tell broker to wait for ack
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// --- Pseudo channel to avoid main() to finish
	forever := make(chan bool)

	// --- Go Routine, fetch messages (1 at a time, see QoS above)
	go func() {
		for d := range msgs {

			// convert []byte to int
			str := string(d.Body)
			fibint, _ := strconv.Atoi(str)

			log.Printf("Received a message: %d", fibint)
			log.Printf("Calculating fib of %d and blocking further execution...", fibint)
			log.Printf("Result: %d", fib(fibint))
			d.Ack(false)
			time.Sleep(1 * time.Second)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever

}
