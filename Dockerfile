FROM golang:latest

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ENTRYPOINT ["go-wrapper", "run"]

COPY *.go /go/src/app/
COPY *.html /go/src/app/
RUN go-wrapper download
RUN go-wrapper install
