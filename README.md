# Overview
## This demo
This demo is a proof of concept I did for a customer. It´s a simple Kubernetes pod based autoscaler which scales the backend pods based on the queue depth of the message broker. You´ll find the following components here:

	- Kubernetes service for RabbitMQ (service discovery, ext. access for mgmt. UI)
	- Kubernetes deployments for RabbitMQ, sender, receiver and autoscaler

The sender just generates random integers in random intervals, publishs them to RabbitMQ, where they´re pulled by *n* (autoscaled) workers which calculate the fibonacci sequence to simulate CPU load. 

![alt tag](https://github.com/embano1/pubsub_autoscaler/blob/master/img/rmq_qdepth_example.png)

## Why not the HPA (horizontal pod autoscaler)? 
I wanted to use a custom metric, i.e. the queue depth of RabbitMQ as a trigger for scaling. Again, proof-of-concept only, code probably also needs some cleaning and can be shortened ,)  
  
## Quick Start
You need a running Kubernetes environment to deploy this demo. 

	- git clone https://github.com/embano1/pubsub_autoscaler.git
	- cd pubsub_autoscaler
	- kubectl create -f examples/
	- (wait for images to be pulled and pods started)
	- Access RabbitMQ UI through NodePort and NodeIP (depends on your environment)
	- Scale sender deployment and see how the system (RabbitMQ queue details) adapts: kubectl scale deploy sender --replicas=20

If you work with minikube, you can get the port of the RabbitMQ management UI with "minikube service list".  
  
## High-level architecture
![alt tag](https://github.com/embano1/pubsub_autoscaler/blob/master/img/high-level_architecture.png)
  
  
# Requirements/ tested configurations
## Software requirements (my test environment)
	- kubectl v1.4.3
	- minikube v0.13.0 or Kubernetes v1.4.6+coreos.0 (from vagrant multi-node https://github.com/coreos/coreos-kubernetes/tree/master/multi-node/vagrant)
	- RabbitMQ 3.6.6


## Software requirements (my build environment)
	- kubectl v1.4.3
	- minikube v0.13.0 **or** Kubernetes v1.4.6+coreos.0 (from vagrant multi-node https://github.com/coreos/coreos-kubernetes/tree/master/multi-node/vagrant)
	- RabbitMQ 3.6.6
	- https://github.com/streadway/amqp v0.9.1
	- https://github.com/kubernetes/client-go (tests and v1.2 based on [unstable] master tree, commit ID 6841809)
	- Built with go1.7.4 darwin/amd64 and Docker v1.12.3

## Build
    - git clone https://github.com/embano1/pubsub_autoscaler.git
    - cd pubsub_autoscaler
    - go get -d ./...
    - make all
    - (push Docker images to you repo and modify deployments to match your image spec)
    - make clean

## Modify autoscaler code
If you want to modify (i.e. improve) the basic autoscaler logic, you can easily test this with [autoscaler_ext](https://github.com/embano1/pubsub_autoscaler/tree/master/cmd/autoscaler_ext). This file let´s you test your autoscaler code quickly, e.g. go run ..., against the Kubernetes cluster **from outside** the cluster. 

At least modify var kubeconfig to point to your kube config file.  

# Other
## Accessing RabbitMQ statistics
	- Get the NodePort of the service: kubectl describe service rabbitmq (for target port **15672** which is the mgmt. interface)
	- Access the management portal through your browser, e.g. http://172.17.4.101:30383/
	- Username/ password: guest/guest

## Get autoscaler metrics
    - kubectl get po | grep autoscaler  
    - kubectl logs -f <autoscaler_pod>

# Cleanup
kubectl delete -f examples/
