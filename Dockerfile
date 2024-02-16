# syntax=docker/dockerfile:1

FROM --platform=${BUILDPLATFORM} golang:1.22 as base
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

FROM --platform=${BUILDPLATFORM} base as build

RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o /usr/local/bin/spin-nats-proxy ./cmd/proxy/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM --platform=${TARGETPLATFORM} gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /usr/local/bin/spin-nats-proxy .
USER 65532:65532

ENTRYPOINT ["/spin-nats-proxy"]
