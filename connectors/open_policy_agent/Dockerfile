FROM golang:1.10 AS build
WORKDIR /go/src
COPY go ./go
COPY main.go .

ENV CGO_ENABLED=0
RUN go get -d -v ./...

RUN go build -a -installsuffix cgo -o open_policy_agent .

FROM scratch AS runtime
ENV GIN_MODE=release
COPY --from=build /go/src/open_policy_agent ./
EXPOSE 8080/tcp
ENTRYPOINT ["./open_policy_agent"]
