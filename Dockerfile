FROM golang:1.13 AS builder
WORKDIR /app
COPY . .
ENV GO111MODULE on
ENV GOOS linux
ENV CGO_ENABLED 0
WORKDIR /app/cmd/server
RUN go build

FROM alpine:latest
RUN apk add --no-cache ca-certificates bind-tools
COPY --from=builder /app/cmd/server /usr/local/bin
EXPOSE 53/udp 25 80 443
CMD ["server"]
