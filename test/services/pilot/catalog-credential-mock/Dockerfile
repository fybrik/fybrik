#FROM golang:1.13.8
FROM golang:1.13.8-alpine

WORKDIR /m4d/
COPY go.mod go.mod

COPY pkg/connectors/protobuf pkg/connectors/protobuf
COPY connectors/vault connectors/vault

WORKDIR /m4d/test/services/pilot/catalog-credential-mock
COPY test/services/pilot/catalog-credential-mock .

RUN go get -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o catalog-credential-mock mockup_pilot_catalog_con.go

EXPOSE 50081

CMD [ "./catalog-credential-mock"]
