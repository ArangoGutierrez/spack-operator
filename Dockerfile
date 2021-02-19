# Build the multiarch-builder-operator binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o multiarch-builder-operator main.go

FROM registry.access.redhat.com/ubi8/ubi 

ARG ASSETS_DIR=build/assets

COPY --from=builder /workspace/multiarch-builder-operator .
#COPY ${ASSETS_DIR} /assets

WORKDIR /
USER multiarch-builder-operator

ENTRYPOINT ["/multiarch-builder-operator"]
