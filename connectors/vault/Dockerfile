#FROM golang:1.13.8
FROM golang:1.13.8-alpine

WORKDIR /m4d/
COPY go.mod go.mod

WORKDIR /m4d/pkg/connectors/
COPY pkg/connectors/protobuf protobuf

WORKDIR /m4d/connectors/vault
COPY connectors/vault .

RUN go get -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o vault-connector *.go

EXPOSE 50083

CMD ["./vault-connector"]