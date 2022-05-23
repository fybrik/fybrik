FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6-751
ENV HOME=/tmp
WORKDIR /tmp

COPY datacatalog /
USER 1001

ENTRYPOINT ["/datacatalog"]
