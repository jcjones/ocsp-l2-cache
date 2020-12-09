FROM golang:1.15-buster as builder
RUN mkdir /build
WORKDIR /build
RUN mkdir bin gopath
ENV GOPATH /build/gopath

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /build/tini
RUN chmod +x /build/tini

ADD . /build/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/ocsp-l2-cache

FROM scratch

COPY --from=builder /build/bin/ocsp-l2-cache /app/
COPY --from=builder /build/tini /

CMD ["/app/ocsp-l2-cache"]

WORKDIR /app

EXPOSE 8080/tcp
EXPOSE 8081/tcp
