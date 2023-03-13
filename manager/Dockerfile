# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
ARG tag=8.7
FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag

ENV HOME=/tmp
WORKDIR $HOME
COPY manager /
USER 10001

ENTRYPOINT ["/manager"]
