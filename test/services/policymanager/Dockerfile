FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6

ENV HOME=/tmp
WORKDIR /tmp

COPY serverpolicymanager /
USER 10001

ENTRYPOINT ["/serverpolicymanager"]
