my_ip=$(docker-machine ip)

docker run -it --rm redis:5 redis-cli -h ${my_ip}