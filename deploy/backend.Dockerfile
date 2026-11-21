# Pin the base images by digest in production, e.g.
#   docker build --build-arg GO_IMAGE=golang:1.25-alpine@sha256:... .
# Keeping them as build args lets operators freeze digests without editing this
# file, while the default stays readable.
ARG GO_IMAGE=golang:1.25-alpine
FROM ${GO_IMAGE} AS build
WORKDIR /src
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
        -ldflags="-s -w -X gopress/internal/buildinfo.Version=${VERSION} -X gopress/internal/buildinfo.Commit=${COMMIT} -X gopress/internal/buildinfo.BuildTime=${BUILD_TIME}" \
        -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
        -ldflags="-s -w -X gopress/internal/buildinfo.Version=${VERSION} -X gopress/internal/buildinfo.Commit=${COMMIT} -X gopress/internal/buildinfo.BuildTime=${BUILD_TIME}" \
        -o /out/seed ./cmd/seed

ARG RUNTIME_IMAGE=alpine:3.21
FROM ${RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates tzdata wget && adduser -D -u 10001 gopress
WORKDIR /app
RUN mkdir -p /app/uploads && chown gopress:gopress /app/uploads
COPY --from=build /out/server /out/seed /usr/local/bin/
COPY migrations ./migrations
USER gopress
EXPOSE 8080
CMD ["sh", "-c", "seed && server"]
