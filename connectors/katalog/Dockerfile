FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6

ENV HOME=/tmp
WORKDIR /tmp

COPY bin/katalog /katalog
USER 1001

ENTRYPOINT ["/katalog"]
CMD [ "run" ]
