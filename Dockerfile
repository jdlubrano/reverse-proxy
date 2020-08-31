FROM golang:1.14-alpine

ENV PORT 8080

RUN apk add --no-cache git

ADD . /go/src/github.com/jdlubrano/reverse-proxy

WORKDIR /go/src/github.com/jdlubrano/reverse-proxy

RUN go mod download && go install -tags musl

ADD . /go/src/github.com/jdlubrano/reverse-proxy

RUN go build

EXPOSE ${PORT}

CMD ["reverse-proxy"]
