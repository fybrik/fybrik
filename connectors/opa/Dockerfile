ARG image=registry.access.redhat.com/ubi8/ubi-minimal:8.7
FROM $image

ENV HOME=/tmp
WORKDIR /tmp

COPY opa-connector /

EXPOSE 8080
USER 10001

ENTRYPOINT ["/opa-connector"]
CMD [ "run" ]
