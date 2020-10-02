#FROM golang:1.13.8
FROM golang:1.13.8-alpine

WORKDIR /m4d/
COPY go.mod go.mod

COPY pkg/connectors/protobuf pkg/connectors/protobuf

WORKDIR /m4d/connectors/opa
COPY connectors/opa .

RUN go get -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o opa-connector *.go

EXPOSE 50082

CMD ["./opa-connector"]
