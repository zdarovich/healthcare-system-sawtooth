# healthcare-system-sawtooth

## Install before build on MacOS
```
brew install zmq
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/opt/openssl/lib/pkgconfig
```

## Run sea from local
```
go run cmd/client/main.go user -n redax -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/redax.priv
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

## Run user_a client
```
go run cmd/client/main.go user -n user_a -u rest-api:8008 -V tcp://validator:4004 -k resources/keys/a.priv -d true -l /ip4/0.0.0.0/tcp/5003 -b /ip4/0.0.0.0/tcp/5002/p2p/16Uiu2HAmDhP9i7H5UD4EBysVrn2NrjPeC5poE134AQFfUFgL3yTG
```

## Run user_b client
```
go run cmd/client/main.go user -n user_b -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/b.priv -d true -b /ip4/0.0.0.0/tcp/5001/p2p/16Uiu2HAmLJ8oWvBLhZAb9uQuXLqKfXBVAJW1rUfi6BEg5e7eXfHj
```

# List shared of user_a
```
ls-shared 64085cb5128f68c6e514e83c70c973d8ef4643736147e5fb83ca308a20671d0db4e241 /
```