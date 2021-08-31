# healthcare-system-sawtooth

## Install before build on MacOS
```
brew install zmq
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/opt/openssl/lib/pkgconfig
```

## Generate admin identity
```
go run cmd/client/main.go generate -n admin -p resources/keys
```
## Run admin identity and register it
```
go run cmd/client/main.go user -n admin -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/admin.priv
```

## Generate user_a identity
```
go run cmd/client/main.go generate
```
## Run user_a identity and register it
```
go run cmd/client/main.go user -n user_a -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/a.priv
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
go run cmd/client/main.go user -n user_a -u localhost-api:8008 -V tcp://localhost:4004 -k resources/keys/a.priv
```

## Run user_b client
```
go run cmd/client/main.go user -n user_b -u localhost:8008 -V tcp://localhost:4004 -k resources/keys/b.priv
```

# List shared of user_a
```
ls-shared 64085cb5128f68c6e514e83c70c973d8ef4643736147e5fb83ca308a20671d0db4e241 /
```

## Run benchmark tests
```
go test -v test/*_test.go
```
