FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6

ENV HOME=/tmp
WORKDIR /tmp

COPY opa-connector /

EXPOSE 8080
USER 1001

ENTRYPOINT ["/opa-connector"]
CMD [ "run" ]
