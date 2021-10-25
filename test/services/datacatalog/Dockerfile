FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV HOME=/tmp
WORKDIR /tmp

COPY datacatalog /
USER 1001

ENTRYPOINT ["/datacatalog"]
