# Start with golang base image
FROM golang:latest
MAINTAINER Omar Qazi (omar@smick.co)

# Compile latest source
ADD . /go/src/github.com/omarqazi/hearst
RUN go get github.com/omarqazi/hearst
RUN go get bitbucket.org/liamstask/goose/cmd/goose
RUN go install github.com/omarqazi/hearst

WORKDIR /go/src/github.com/omarqazi/hearst
ENTRYPOINT /go/bin/hearst
EXPOSE 8080
