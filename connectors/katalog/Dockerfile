FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY bin/katalog .
USER nonroot:nonroot

ENTRYPOINT ["/katalog"]
CMD [ "run" ]
