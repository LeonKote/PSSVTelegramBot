# Builder
FROM golang:1.24-alpine AS builder
RUN apk add build-base yq && \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
ADD . /src/service
WORKDIR /src/service
RUN make
WORKDIR /src/service/cmd/cameras
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-s -w' -o /build/service .

# Service
FROM alpine:latest
RUN apk add ffmpeg
COPY --from=builder /build/service /usr/bin
ENTRYPOINT ["/usr/bin/service"]