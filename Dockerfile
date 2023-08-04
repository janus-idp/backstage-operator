# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset:latest as builder
USER 0

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
# COPY api/ api/
# COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM registry.access.redhat.com/ubi9/ubi-micro:latest

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
