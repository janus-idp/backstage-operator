# Build the manager binary
#@follow_tag(registry.redhat.io/rhel9/go-toolset:1.19)
FROM registry.access.redhat.com/ubi9/go-toolset:1.19 as builder
USER 0
ENV GOPATH=/go/

# Upstream sources
# Downstream comment
ENV EXTERNAL_SOURCE=.
ENV CONTAINER_SOURCE=/opt/app-root/src
WORKDIR /workspace
#/ Downstream comment

# Downstream sources
# Downstream uncomment
# ENV EXTERNAL_SOURCE=$REMOTE_SOURCES/upstream1/app/distgit/containers/rhdh-operator
# ENV CONTAINER_SOURCE=$REMOTE_SOURCES_DIR
# WORKDIR $CONTAINER_SOURCE/
#/ Downstream uncomment

COPY $EXTERNAL_SOURCE ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
# Downstream comment
RUN go mod download
#/ Downstream comment

# Downstream uncomment
# COPY $REMOTE_SOURCES/upstream1/cachito.env ./
# RUN source ./cachito.env && rm -f ./cachito.env && mkdir -p /workspace
#/ Downstream uncomment

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o /workspace/manager main.go && ls -la /workspace/

# NOTE: ubi-micro is not be FIPS compliant, if openssl is not installed
#@follow_tag(registry.redhat.io/ubi9/ubi-micro:9.2)
FROM registry.access.redhat.com/ubi9/ubi-micro:9.2

ENV HOME=/opt/helm \
    USER_NAME=helm \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

# Copy necessary files with the right permissions
COPY --chown=${USER_UID}:0 watches.yaml ${HOME}/watches.yaml
COPY --chown=${USER_UID}:0 helm-backstage  ${HOME}/helm-backstage

# Copy manager binary
COPY --from=builder /workspace/manager .

USER ${USER_UID}

WORKDIR ${HOME}

ENTRYPOINT ["/manager"]

# append Brew metadata here
