ARG tag=8.7
FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag

ENV HOME=/tmp
WORKDIR /tmp

COPY bin/katalog /katalog
USER 10001

ENTRYPOINT ["/katalog"]
CMD [ "run" ]
