# Copyright 2022 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

ARG tag=8.7
FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag
COPY solver-tools/* /hack/tools/
RUN microdnf install tar gzip
RUN /hack/tools/install_or_tools.sh
RUN rm /hack/tools/*.sh

FROM registry.access.redhat.com/ubi8/ubi-minimal:$tag
ENV HOME=/data
WORKDIR /data/tools
COPY --from=0 /hack/tools .
USER 1001
