ARG go_version
FROM golang:${go_version} AS builder
WORKDIR /app
COPY . .
ENV GO111MODULE on
ENV GOOS linux
ENV CGO_ENABLED 0
WORKDIR /app
RUN make build-server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /usr/local/bin
EXPOSE 53/udp 25 80 443
CMD ["server"]
