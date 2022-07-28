FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV HOME=/tmp
WORKDIR /tmp

COPY bin/katalog /katalog
USER 10001

ENTRYPOINT ["/katalog"]
CMD [ "run" ]
