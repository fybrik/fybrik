FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV HOME=/tmp
WORKDIR /tmp

COPY catalog-credential-mock /
USER 1001

ENTRYPOINT ["/catalog-credential-mock"]
