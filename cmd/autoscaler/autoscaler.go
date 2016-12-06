/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	// RabbitMQ
	"github.com/streadway/amqp"

	// Kubernetes
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

// error handling
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// watch rabbitmq channel metrics (needed for scale function)
func inspect(ch *amqp.Channel, qinspectch chan int) {
	for {
		q, err := ch.QueueInspect("fib-work")
		failOnError(err, "Failed to open a channel")
		// log.Printf("Queue: %v, Messages: %v, Consumers: %v", q.Name, q.Messages, q.Consumers)
		qinspectch <- q.Messages
		time.Sleep(1 * time.Second)
	}
}

// scale deployment based on number of messages in the rabbitmq channel
func scale(clientset *kubernetes.Clientset, namespace *string, deployment *string, qmax int, qmin int, scalech chan int, qinspectch chan int) {
	var replicas int32
	var lastval int
	var qdepth int

	for {

		// --- Get queue depth
		lastval = qdepth
		qdepth = <-qinspectch
		log.Printf("Currently %d messages in queue (last seen: %d)", qdepth, lastval)

		// --- Get K8s deployment information (replicas)
		// v1beta1.ScaleSpec.replicas
		// desired (here: current) number of instances for the scaled object
		currentscale, err := clientset.Scales(*namespace).Get("Deployment", *deployment)
		if err != nil {
			panic(err.Error())
		}

		replicas = currentscale.Spec.Replicas
		//log.Printf("Current qty of replicas for deployment %v: %v\n", *deployment, currentscale.Spec.Replicas)

		// --- Scale up/ down pods
		switch {
		case qdepth > qmax && qdepth > lastval:

			replicas++
			// --- Update K8s API .Spec with target state (replicas)
			currentscale.Spec = v1beta1.ScaleSpec{Replicas: replicas}

			// --- Scale up deployment
			_, err = clientset.Scales(*namespace).Update("Deployment", currentscale)
			if err != nil {
				panic(err.Error())
			}
			log.Printf("Scaled %v up to qty: %d\n", *deployment, replicas)

		case qdepth > qmin && qdepth < qmax && currentscale.Spec.Replicas > 1:
			replicas--
			//fmt.Println(replicas)

			// --- Update K8s API .Spec with target state (replicas)
			currentscale.Spec = v1beta1.ScaleSpec{Replicas: replicas}

			// --- Scale down deployment
			_, err = clientset.Scales(*namespace).Update("Deployment", currentscale)
			if err != nil {
				panic(err.Error())
			}
			log.Printf("Scaled %v down to qty: %d\n", *deployment, replicas)

		case qdepth < qmin && currentscale.Spec.Replicas >= 1:
			replicas = 0
			//fmt.Println(replicas)

			// --- Update K8s API .Spec with target state (replicas)
			currentscale.Spec = v1beta1.ScaleSpec{Replicas: replicas}

			// --- Scale down deployment
			_, err = clientset.Scales(*namespace).Update("Deployment", currentscale)
			if err != nil {
				panic(err.Error())
			}
			log.Printf("Queue below watermark (qmin: %d, qmax: %d)...reducing to %d replicas\n", qmin, qmax, replicas)
		}

		time.Sleep(10 * time.Second)
	}
}

func main() {

	// --- Declare channels for program synchronization
	qinspectch := make(chan int)
	scalech := make(chan int)
	forever := make(chan bool)

	// --- Command line options
	brokerptr := flag.String("b", "rabbitmq", "SVC (Kubernetes) name where to find the RabbitMQ broker (can also be an IP)")
	portptr := flag.String("p", "5672", "Port where RabbitMQ broker listens on")
	namespace := flag.String("ns", "default", "Use this namespace")
	deployment := flag.String("d", "receiver", "Scale this deployment")
	qmaxptr := flag.Int("qmax", 100, "Upper queue watermark before starting scale operations")
	qminptr := flag.Int("qmin", 20, "Lower queue watermark to scale down to")
	flag.Parse()

	// Dereference some pointers
	broker := *brokerptr
	port := *portptr
	qmax := *qmaxptr
	qmin := *qminptr

	// --- RabbitMQ initialization
	// get a connection
	conn, err := amqp.Dial("amqp://guest:guest@" + broker + ":" + port)
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Println("Connected")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	log.Println("Opened a channel")
	defer ch.Close()

	// --- init a k8s go-client context
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	failOnError(err, "Failed to create config in cluster.")
	log.Println("Created in-cluster config")

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	failOnError(err, "Failed to create client set.")
	log.Println("Created client set")

	// --- Program
	go inspect(ch, qinspectch)
	go scale(clientset, namespace, deployment, qmax, qmin, scalech, qinspectch)

	// --- Don´t quit main() by infinitely waiting on this channel (no values send to)
	<-forever

}
