FROM registry.access.redhat.com/ubi8/ubi-minimal
ENV HOME=/tmp
WORKDIR /tmp

COPY manager-client /

EXPOSE 50082
USER 1001

ENTRYPOINT ["/manager-client"]
