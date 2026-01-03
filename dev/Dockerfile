FROM golang:latest
ENV LOCAL_BIN=/go/bin
ENV PATH="/app/dist:${PATH}"
COPY Makefile .
RUN make tools && rm Makefile
