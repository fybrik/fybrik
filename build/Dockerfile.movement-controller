FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
