FROM golang:latest
ENV PATH="/app/build:${PATH}"
COPY Makefile .
RUN make devtools && rm Makefile
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash
RUN apt-get update && apt-get install -y nodejs
