FROM golang:1.15-buster as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN mkdir bin gopath
ENV GOPATH /build/gopath

RUN go build -o bin/ocsp-l2-cache

FROM python:3.8-buster
RUN apt update && apt install -y ca-certificates && \
    apt -y upgrade && apt-get autoremove --purge -y && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

RUN adduser --system --uid 10001 --group --home /app app

COPY --from=builder /build/bin /app/
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "-g", "--", "/app/ocsp-l2-cache"]

USER app
WORKDIR /app

EXPOSE 80/tcp
EXPOSE 8080/tcp