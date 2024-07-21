ARG go_version=1.19

FROM golang:${go_version} AS builder
WORKDIR /opt/app
COPY . .
ENV GO111MODULE on
ENV GOOS linux
ENV CGO_ENABLED 0
WORKDIR /opt/app
RUN make build-server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /opt/app
COPY --from=builder /opt/app/server .
EXPOSE 53/udp 21 25 80 443
CMD ["./server", "serve"]
