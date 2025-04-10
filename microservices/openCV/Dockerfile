# Builder
FROM golang:1.24-alpine3.21 AS builder
RUN apk update && apk add build-base yq pkgconfig opencv-dev=4.10.0-r3
RUN apk add 
ADD . /src/service
WORKDIR /src/service/cmd/openCV
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-s -w' -o /build/service .

# Service
FROM alpine:3.21
RUN apk update && apk add ffmpeg opencv-dev=4.10.0-r3
COPY --from=builder /build/service /usr/bin
ENTRYPOINT ["/usr/bin/service"]