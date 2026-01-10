FROM golang:latest
ENV PATH="/app/build:${PATH}"
RUN apt-get update && apt-get install -y \
  bind9-host \
  curl \
  swaks \
  ncftp \
  && rm -rf /var/lib/apt/lists/*
COPY Makefile .
RUN make devtools && rm Makefile
