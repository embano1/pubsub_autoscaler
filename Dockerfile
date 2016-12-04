FROM scratch
MAINTAINER Michael Gasch <michael_gasch@live.com>
ADD cmd/receiver/receiver /receiver
ADD cmd/sender/sender /sender
ADD cmd/autoscaler/autoscaler /autoscaler
