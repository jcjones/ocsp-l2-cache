# OCSP-L2-Cache: A Caching OCSP Forwarder

This OCSP responder is intended to lie between an authoritative OCSP responder and clients, caching responses in a Redis key-value store to reduce load on the authoritative responder.

## Requirements

Redis 5

## Configuration

TBD. Currently edit `main.go`

## Interacting

Probably easiest to use tools that can override the responder URL, like OpenSSL or [jcjones/ocspchecker](https://github.com/jcjones/ocspchecker) (assuming the cache is running on `localhost:9020`):

```
go get github.com/jcjones/ocspchecker

ocspchecker -nostaple -responder http://localhost:9020 -url https://letsencrypt.org -dump
```

## TODOs

- [ ] Switch from go-metrics to prometheus
- [ ] Actual configuration mechanism
- [ ] Containers
- [ ] Actually compress the data in `compressedresponse`
- [ ] Don't store the whole headers, synthesize everything we can to reduce storage needs
- [ ] Link-failure tests
- [ ] OcspStore tests with the mock cache
- ... more in the issues
