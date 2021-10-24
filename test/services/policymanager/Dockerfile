FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV HOME=/tmp
WORKDIR /tmp

COPY serverpolicymanager /
USER 1001

ENTRYPOINT ["/serverpolicymanager"]
