ARG BASE_IMAGE=rockylinux/rockylinux:9-ubi-micro

# Compile once per target architecture. BuildKit provides TARGETOS/TARGETARCH
# for each value passed to `docker buildx build --platform ...`.
FROM --platform=$BUILDPLATFORM golang:1.25 AS build

WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /diagnostic-service ./cmd

# Prepare the final image filesystem while running on the build platform. This
# avoids executing the target platform's /bin/sh when building, for example,
# arm64 images on an amd64 builder without QEMU/binfmt support.
RUN mkdir -p /image-root/etc/diagnostic-service \
    && cp /diagnostic-service /image-root/diagnostic-service

FROM ${BASE_IMAGE}
COPY --from=build /image-root/ /
EXPOSE 8080
ENTRYPOINT ["/diagnostic-service"]

# For test
CMD ["-http-addr", "0.0.0.0:8080"]
# CMD ["-config", "/etc/diagnostic-service/config.toml", "-http-addr", "0.0.0.0:8080"]
