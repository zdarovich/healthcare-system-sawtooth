# healthcare-system-sawtooth

## Install before build on MacOS
```
brew install zmq
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/opt/openssl/lib/pkgconfig
```

## Run sea from local
```
go run cmd/client/main.go sea -n redax -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/redax.priv -d true -b /ip4/0.0.0.0/tcp/5001/p2p/16Uiu2HAkwxu3JAoqZ7QQ343hQuADCbkqfimCNRTnqQgoUpvoKEty
```

## Run go client from local
```
go run cmd/client/main.go -c "cmd/client/config.json"
```
## Delete all images, containers, volumes
```
docker ps | awk '{print $1}' | xargs docker stop
docker rm -vf $(docker ps -a -q)
docker rmi -f $(docker images -a -q)
```

## Run healthcare-system in dev mode
```
docker-compose -f docker/sawtooth-default.yaml up -d
```

## Stop and remove 'hyperledger' docker containers
```
docker ps | grep hyperledger | awk '{print $1}' | xargs docker stop
docker container rm $(docker ps --filter "status=exited" | grep 'hyperledger' | awk '{print $1}')
docker rmi $(docker images | grep 'hyperledger' | awk '{print $3}')
```

## View docker logs
```
docker logs --follow CONTAINER_ID
```