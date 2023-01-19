########################################
# Common arguments
ARG OSTOOL_DIR=/usr/local/bin

########################################
# Check required arguments up front (and only query dockerhub once.)
FROM registry.access.redhat.com/ubi8/ubi:latest AS base


########################################
# Install kubectl, tkn, helm, & oc in /usr/local/bin to copy into the dev image.
FROM base AS ostools
USER 0:0

ARG ARCH=x86_64
ARG WARCH=amd64
ARG OSTOOL_DIR

########################################
# We'll create the build environment
FROM base AS dev
ARG OSTOOL_DIR
USER 0:0


# Set up the entrypoint
COPY entrypoint.sh main.sh /
RUN mkdir /entrypoint.d
ENTRYPOINT [ "/entrypoint.sh" ]
CMD []

# Streams requires utf-8 (Force a reliable sort order with LC_COLLATE)
ENV LANG="en_US.UTF-8"
ENV LC_CTYPE="en_US.UTF-8"
ENV LC_NUMERIC="en_US.UTF-8"
ENV LC_TIME="en_US.UTF-8"
ENV LC_COLLATE=C
ENV LC_MONETARY="en_US.UTF-8"
ENV LC_MESSAGES="en_US.UTF-8"
ENV LC_PAPER="en_US.UTF-8"
ENV LC_NAME="en_US.UTF-8"
ENV LC_ADDRESS="en_US.UTF-8"
ENV LC_TELEPHONE="en_US.UTF-8"
ENV LC_MEASUREMENT="en_US.UTF-8"
ENV LC_IDENTIFICATION="en_US.UTF-8"

# Install necessary RPMs for our builds
RUN dnf install -y --disableplugin=subscription-manager         \
                   --setopt=skip_missing_names_on_install=False \
                   --allowerasing                               \
                glibc-langpack-en \
                go-toolset        \
                golang            \
                golang-bin        \
                golang-docs       \
                golang-misc       \
                jq                \
                make              \
                maven             \
                wget              \
                unzip             \
                openssl           \
                && dnf -y clean all \
                && rm -rf /var/cache /var/log/dnf* /var/log/yum.*

RUN dnf install -y --disableplugin=subscription-manager      \
                --setopt=skip_missing_names_on_install=False \
                python38                                     \
                && dnf -y clean all \
                && rm -rf /var/cache /var/log/dnf* /var/log/yum.*

COPY centos.repo /etc/yum.repos.d/centos.repo
COPY RPM-GPG-KEY-CentOS-Official /etc/pki/rpm-gpg/

RUN chmod 1777 /home

RUN wget -nv https://dl.google.com/go/go1.18.9.linux-amd64.tar.gz && \
    tar -xzf go1.18.9.linux-amd64.tar.gz && \ 
    mv go /usr/local && \
    rm go1.18.9.linux-amd64.tar.gz 

RUN cd /tmp/ && \
    wget -nv --no-check-certificate -O ~/SACUtil.zip https://cloud.appscan.com/api/SCX/StaticAnalyzer/SAClientUtil?os=linux && \
    unzip -q ~/SACUtil.zip && \
    dir=$(find ./ -name "SAClientUtil*") && \
    mv $dir /usr/local/bin/SAClientUtil && \
    rm ~/SACUtil.zip

RUN curl -kLo yq https://github.com/mikefarah/yq/releases/download/3.4.0/yq_linux_amd64 \
 && chmod 755 yq && mv yq /usr/local/bin/yq \
 && curl -kLo yq4 https://github.com/mikefarah/yq/releases/download/v4.20.1/yq_linux_amd64 \
 && chmod 755 yq4 && mv yq4 /usr/local/bin/yq4

RUN curl -sSLo - https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v4.5.5/kustomize_v4.5.5_linux_amd64.tar.gz | \
        tar xzf - -C /usr/local/bin/

RUN yum install -y yum-utils && yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && yum -y install docker-ce-cli --allowerasing
    
ENV GOROOT=/usr/local/go 
ENV GOPATH=$HOME/Projects/Proj1 
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH:/usr/local/bin/SAClientUtil/bin/
