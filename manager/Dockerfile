# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
ARG tag=8.7
FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag
# this is required for FIPS support
RUN microdnf -y update && microdnf clean all
RUN microdnf install -y go-toolset
ENV HOME=/tmp
WORKDIR $HOME
COPY manager /
USER 10001

ENTRYPOINT ["/manager"]
