FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY katalog .
USER nonroot:nonroot

ENTRYPOINT ["/katalog"]
CMD [ "run" ]
