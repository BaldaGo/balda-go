FROM ubuntu:16.04
MAINTAINER Nikita Boyarskikh <N02@yandex.ru>

RUN \
  apt-get update && \
  apt-get -y install \
    git \
    golang-go

ENV GOPATH /code

RUN go get -u github.com/BaldaGo/balda-go
RUN cat /code/src/github.com/BaldaGo/balda-go/requirements.txt | xargs go get -u
RUN go build -o /code/src/github.com/BaldaGo/balda-go/balda-go /code/src/github.com/BaldaGo/balda-go/main.go
