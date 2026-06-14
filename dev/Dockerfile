FROM golang:latest
ENV PATH="/app/build:${PATH}"
RUN apt-get update && apt-get install -y \
  bash-completion \
  bind9-host \
  dnsutils \
  net-tools \
  iproute2 \
  iputils-ping \
  curl \
  swaks \
  ncftp \
  && rm -rf /var/lib/apt/lists/*
COPY Makefile .
RUN make devtools && rm Makefile
