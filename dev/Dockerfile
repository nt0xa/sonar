FROM golang:latest
ENV LOCAL_BIN=/go/bin
ENV PATH="/app/build:${PATH}"
COPY Makefile .
RUN make tools && rm Makefile
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash
RUN apt-get update && apt-get install -y nodejs
