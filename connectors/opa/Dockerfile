ARG tag=8.7
FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag

ENV HOME=/tmp
WORKDIR /tmp

COPY opa-connector /

EXPOSE 8080
USER 10001

ENTRYPOINT ["/opa-connector"]
CMD [ "run" ]
