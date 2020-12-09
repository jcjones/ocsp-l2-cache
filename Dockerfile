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

ENV Responders "A84A6A63047DDDBAE6D139B7A64565EFF3A8ECA1=http://ocsp.int-x3.letsencrypt.org;C5B1AB4E4CB1CD6430937EC1849905ABE603E225=http://ocsp.int-x4.letsencrypt.org;142EB317B75856CBAE500940E61FAF9D8B14C2C6=http://r3.o.lencr.org;369D3EE0B140F6272C7CBF8D9D318AF654A64626=http://r4.o.lencr.org"


COPY --from=builder /build/bin/ocsp-l2-cache /app/
COPY --from=builder /build/tini /

CMD ["/app/ocsp-l2-cache"]

WORKDIR /app

EXPOSE 8080/tcp
EXPOSE 8081/tcp
