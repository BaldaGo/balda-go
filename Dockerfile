FROM golang:1.8
MAINTAINER Nikita Boyarskikh <N02@yandex.ru>

RUN \
  apt-get update && \
  apt-get -y install \
    git \
    golang-go


ENV GOPATH /code
ADD . /code/src/github.com/BaldaGo/balda-go
WORKDIR /code/src/github.com/BaldaGo/balda-go
RUN cat requirements.txt | xargs go get -u 

RUN go build
