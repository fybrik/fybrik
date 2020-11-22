FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY serverpolicycompiler .
USER nonroot:nonroot

ENTRYPOINT ["/serverpolicycompiler"]
