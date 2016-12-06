REPO=embano1
BINARY=pubsub_autoscaler
RECEIVER=./cmd/receiver
SENDER=./cmd/sender
AUTOSCALER=./cmd/autoscaler
VERSION=1.2

all: image

build:
	cd ${SENDER}/ && GOOS=linux go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo .
	cd ${RECEIVER}/ && GOOS=linux go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo .
	cd ${AUTOSCALER}/ && GOOS=linux go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo .

image: build
	docker build -t ${REPO}/${BINARY}:${VERSION} .

clean: 
# wont remove the docker image
	rm ${AUTOSCALER}/autoscaler
	rm ${SENDER}/sender
	rm ${RECEIVER}/receiver
	
