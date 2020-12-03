my_ip=$(ipconfig getifaddr en0)

export RedisHost=${my_ip}:6379

go test ./...