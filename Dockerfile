FROM golang:1.12
COPY . /go/src/github.com/moul/ssh2docker
WORKDIR /go/src/github.com/moul/ssh2docker
RUN make
ENTRYPOINT ["/go/src/github.com/moul/ssh2docker/ssh2docker"]
