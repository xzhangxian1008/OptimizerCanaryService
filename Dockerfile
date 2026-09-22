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

FROM ${BASE_IMAGE}
COPY --from=build /diagnostic-service /diagnostic-service
EXPOSE 8080
ENTRYPOINT ["/diagnostic-service"]

# For test
CMD ["-http-addr", "0.0.0.0:8080"]
