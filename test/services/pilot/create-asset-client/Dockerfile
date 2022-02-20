FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV HOME=/tmp
WORKDIR /tmp

COPY create-asset-client /
USER 1001

ENTRYPOINT ["/create-asset-client"]
