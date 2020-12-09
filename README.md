# OCSP-L2-Cache: A Caching OCSP Forwarder

This OCSP responder is intended to lie between an authoritative OCSP responder and clients, caching responses in a Redis key-value store to reduce load on the authoritative responder.

## Requirements

Redis 5

## Configuration

Set these environment variables:
* ListenOCSP
  - default: `:8080`
* ListenHealth
  - default: `:8081`
* RedisHost
  - default: `redis:6379`
* ID
  - default: hostname
* SyslogProto
  - default: `""`
  - type: udp, tcp, or blank for local socket
* SyslogAddr
  - default: `""`
* CacheLifespan
  - default: `24h`
  - type: [Duration](https://golang.org/pkg/time/#ParseDuration)
* ConnectionDeadline
  - default: `1s`
  - type: [Duration](https://golang.org/pkg/time/#ParseDuration)
* Responders
  - type: `key ID in hex=http://url;...`

Example run:

```
Responders="A84A6A63047DDDBAE6D139B7A64565EFF3A8ECA1=http://ocsp.int-x3.letsencrypt.org;C5B1AB4E4CB1CD6430937EC1849905ABE603E225=http://ocsp.int-x4.letsencrypt.org;142EB317B75856CBAE500940E61FAF9D8B14C2C6=http://r3.o.lencr.org;369D3EE0B140F6272C7CBF8D9D318AF654A64626=http://r4.o.lencr.org" go run main.go
```

An arbitrary number of these l2-cache instances can point to a Redis cluster for horizontal scaling. Once you run into issues at the Redis cluster, you can just construct another whole cluster.

## Interacting

Probably easiest to use tools that can override the responder URL, like OpenSSL or [jcjones/ocspchecker](https://github.com/jcjones/ocspchecker) (assuming the cache is running on `localhost:9020`):

```
go get github.com/jcjones/ocspchecker

ocspchecker -nostaple -responder http://localhost:9020 -url https://letsencrypt.org -dump
```

## Building and running

Via Docker:

```
docker build -t ocsp-l2-cache .

docker run --rm ocsp-l2-cache --publish 8080:80 --publish 8081:8080
```


## TODOs

- [ ] Switch from go-metrics to prometheus
- [ ] Actual configuration mechanism
- [ ] Containers
- [ ] Actually compress the data in `compressedresponse`
- [ ] Don't store the whole headers, synthesize everything we can to reduce storage needs
- [ ] Link-failure tests
- [ ] OcspStore tests with the mock cache
- [ ] Admin API interface for pushing new cache entries, flushing entries
- [ ] Deployment guidance
- ... more in the issues
