# Healthcare system on Hyperledger Sawtooth
## Introduction
Healthcare system decentralized data storage.
Encrypted keys are stored on the Hyperledger Sawtooth blockchain.
Encrypted data is stored on the off-chain using MongoDB.
Blockhain is accessed using transaction proccessor, which stores user data, user shared data, encrypted keys.
Storage is managed by the client, which fetches data from MongoDB and decrypts cipher using data from the Hyperledger Sawtooth blockchain.
All components are deloyed in the Docker containers

## Prerequisites
- Docker

## Directories and files
- `/client`: client, which encrypts, decrypts, communicates with the blockchain
- `/cmd`: commands to run interactive prompt
- `/crypto`: libraries for encryption/decryption
- `/docker`: docker infrastructure files
- `/resources`: pre-built private and public keys for quick testing
- `/test`: benchmark tests
- `/tp`: smart contract, which stores users and the data
- `go.mod`: list of Golang libraries used in the project
- `go.sum`: hash sums of Golang libraries

### Client commands description
- `register`: Register current identity as user on the blockchain.
- `sync`: Sync data from the blockchain.
- `whoami`: Get current user info.
- `create <data_name> <data>`: Create encrypted data on the blockchain and store it off-chain.
- `share <hash> <username>`: Share own data to other user by hash and user to share with username.
- `ls`: List all data owned by current user on the blockchain.
- `get <hash>`: Get own data by hash
- `ls-users`: List all users on the blockchain.
- `ls-shared <username>`: List all shared data by user.
- `get-shared <hash> <username>`: Get shared data by hash and username
- `exit`: Exit command prompt.


## Run and test healthcare system
### Start system
```
docker-compose -f docker/sawtooth-default.yaml up -d
```
### Run command client help
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -h
```
### Run admin identity and register it
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n admin -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/admin.priv
```
### Run a identity and register it
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n a -u rest-api-1:8008 -V tcp://validator-1:4004 -k /app/resources/keys/a.priv
```
### Run b identity and register it
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n b -u rest-api-1:8008 -V tcp://validator-1:4004 -k /app/resources/keys/b.priv
```
### Run benchmark tests
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client go test -v test/*_test.go
```

### Benchmark output description
- `Cumulative`: Aggregate of all sample durations.
- `HMean`: Event duration harmonic mean.
- `Avg.`: Average event duration per sample.
- `p<N>`: Nth %ile.
- `Long 5%`: Average event duration of the longest 5%.
- `Short 5%`: Average event duration of the shortest 5%.
- `Max`: Max observed event duration.
- `Min`: Min observed event duration.
- `Range`: The delta between the max and min sample time
- `StdDev`: The population standard deviation
- `Rate/sec.`: Per-second rate based on cumulative time and sample count.

## Example. How to access shared data. Step by step.
1. Run help from prompt to get information about possible commands
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -h
```
2. Start and register 'admin' identity in the first place because Sawtooth blockchain requires admin user identity
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n admin -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/admin.priv
```
3. Choose exit in the prompt. To force exit use Ctrl + C
4. Start and register 'a' identity
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n a -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/a.priv
```
5. Choose exit in the prompt. To force exit use Ctrl + C
6. Start and register 'b' identity
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n b -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/b.priv
```
7. Write command to create data on the blockchain. First argument is the name of the data. Second argument is the data itself. Data can be string of any length and format.
```
create data1 <data>
```
8. List data owned by current user and get it by hash
```
ls
get fsdfsfsdfsd
```
9. List users on the blockchain
```
ls-users
```
10. Take data hash from 'ls' command, take name of user from 'ls-users' command. Execute command to share data.
First argument takes hash of the data, second argument accepts username.
```
share fsdfsfsdfsd a
```
11. Choose exit in the prompt. To force exit use Ctrl + C
12. Start 'a' identity
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client /app/main user -n a -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/a.priv
```
13. List shared data by username. First argument accepts username.
```
ls-shared b
```
14. Get shared data by hash. First argument accepts hash, second argument accepts username
```
get-shared fsdfsfsdfsd b
```


## Utility commands
### Delete all images, containers, volumes
```
docker ps | grep 'hyperledger' | awk '{print $1}' | xargs docker stop
docker ps | grep 'docker_' | awk '{print $1}' | xargs docker stop
docker container rm $(docker ps --filter "status=exited" | grep 'hyperledger' | awk '{print $1}')
docker container rm $(docker ps --filter "status=exited" | grep 'docker_' | awk '{print $1}')
docker volume rm docker_poet-shared
docker rmi $(docker images | grep 'hyperledger' | awk '{print $3}') --force
docker rmi $(docker images | grep 'docker_' | awk '{print $3}') --force
```

### View docker logs
```
docker logs --follow <CONTAINER_ID>
```

### Install before build on MacOS
```
brew install zmq
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/local/opt/openssl/lib/pkgconfig
```

## NOTE
- Use MongoDB Compass to explore data stored in the MongoDB