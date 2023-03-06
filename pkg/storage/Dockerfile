FROM registry.access.redhat.com/ubi8/ubi-minimal:8.7

ENV HOME=/tmp
WORKDIR /tmp

COPY bin/storagemanager /storagemanager
USER 10001

ENTRYPOINT ["/storagemanager"]
CMD [ "run" ]
