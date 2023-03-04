# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
ARG image=registry.access.redhat.com/ubi8/ubi-minimal:8.7
FROM $image
RUN microdnf -y update && microdnf clean all
# this is required for FIPS support
RUN microdnf install -y go-toolset
ENV HOME=/tmp
WORKDIR $HOME
COPY manager /
USER 10001

ENTRYPOINT ["/manager"]
