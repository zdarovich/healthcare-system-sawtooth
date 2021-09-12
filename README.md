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
- `request-as-third-party <request_from> <data_of_user> <emergency_condition>`: Request data of patient from trusted party as third party
- `request-as-trusted-party <request_from>`: Request data of patient as trusted party
- `list-requests`: List of data requests received from users
- `process-request <OID> <true/false>`: List of data requests received from users
- `batch-upload <path_to_csv_file>`: Upload patient's data using csv file format
- `exit`: Exit command prompt.

### Batch file csv format description
Columns
- `data_name`: The name of the data stored stored in the system
- `trusted_party`: Trusted party. Identity with whom the data is shared. Can be an array of trusted party usernames separated by space. Example: doctorA doctorB
- `access_type`: Emergency condition when data is allowed to be shared to third parties
    0 - Unset. Data cannot be shared to third parties.
    1 - Regular. Data can be shared in regular emergency case.
    2 - Critical. Data can be shared in critical emergency cases. It also includes regular cases.

Example
```csv
name,surname,Blood type,trusted_party,access_type
John,Sally,positive,doctorA,1
John,Sally,positive,doctorA doctorB,1
```

## Run and test healthcare system
### Start the system
```
docker-compose -f docker/sawtooth-default.yaml up -d
```
### Register admin identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-admin /app/main user -n admin -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/admin.priv
exit
```
### Run benchmark tests
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client-admin go test -v test/*_test.go
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


## Example.
- 'patientA' uploads multiple data to the system using csv data upload.
- Data is shared with trusted party 'doctorA'.
- 'thirdPartyA' requests access for data of 'patientA' from 'doctorA'
- 'doctorA' accepts request. Data of 'patientA' is shared by 'doctorA' to 'thirdPartyA' for 1 minute

1. Start and register 'doctorA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-doctor-a /app/main user -n doctorA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/doctorA.priv
exit
```
1.1. Start and register 'doctorB' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-doctor-b /app/main user -n doctorB -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/doctorB.priv
exit
```

2. Start and register 'patientA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-a /app/main user -n patientA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/patientA.priv
```

3. Batch upload 'patientA' data from csv file '/resources/data/patientA.csv'
Commands
```
batch-upload <path_to_csv_file>
```
Example
```
batch-upload /resources/data/patientA.csv
exit
```

4. Start and register 'thirdPartyA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-thirdparty-a /app/main user -n thirdPartyA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/thirdPartyA.priv
```

5. Request 'patientA' data from 'doctorA' as 'thirdPartyA' identity with 'regular' access type
Commands
```
request-as-third-party <request_from> <username_to_take_data_from> <emergency_case>
```
Example
```
request-as-third-party doctorA patientA 1
exit
```

6. Start 'doctorA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-doctor-a /app/main user -n doctorA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/doctorA.priv
```

7. List data requests for 'doctorA' identity
```
list-requests
```
Example response from 'list-requests' command
```json
{
        "OID": "61360d314de12856ad923de3",
        "hash": "88fbd28a1b80eb5966120f1f68fd94bbe58d84adfe510787b1c6f82a61a1b9177da4d137139cf244ec39237995e8044e65927fd808ae2a2539d919fa9f620956",
        "name": "",
        "request_from": "doctorA",
        "username_from": "patientA",
        "username_to": "thirdPartyA",
        "status": 0,
        "access_type": 1
}
```

8. Accept request by 'doctorA' identity to provide data access to 'thirdPartyA' identity. Provide 'OID' from 'list-requests' response. Provide 'true' or 'false' to accept or reject request.
Command
```
process-request <OID> <true/false>
```
Example
```
process-request 61360d314de12856ad923de3 true
exit
```

9. Start 'thirdPartyA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-thirdparty-a /app/main user -n thirdPartyA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/thirdPartyA.priv
```

10. List shared files from 'doctorA'
Commands
```
ls-shared <username>
```
Example
```
ls-shared doctorA
```
Example response from 'ls-shared' command
```json
{
        "Name": "shared_by_doctorA_shared_by_patientA_Name",
        "Hash": "52f220d78ca9cf0da98579f677eb95ef060ca753628faaae325c9e309307e0c824d8e636a05c681b0f29351bf52a5d0da3f217d07a6e9b4f6448d6002f61d940",
        "Size": 4,
        "KeyIndex": "ebec5793147e9663cb4c3ecf3fa65f07909cf8eb3a197a8f949b135b786422274c32e2bebd3b83b384e83af1d133540868aac380047664c9810ad7b8fa475148",
        "Addr": "thirdPartyA",
        "AccessType": 0
}
```

11. Access shared data of 'patientA' from 'doctorA' as 'thirdPartyA' identity
Command
```
get-shared <hash> <username_who_shared_data>
```
Example
```
get-shared 52f220d78ca9cf0da98579f677eb95ef060ca753628faaae325c9e309307e0c824d8e636a05c681b0f29351bf52a5d0da3f217d07a6e9b4f6448d6002f61d940 doctorA
```

## Example. How to add new data sample?
- Place data sample csv file into 'resources/data' directory
Example:
```
resources/data/example_a.csv
```
- Run identity of the user, whom this data belongs to. Run batch-upload command with path starting as '/app/resources/data/<file_name>'
Command
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-a /app/main user -n <identity_username> -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/<identity_key>.priv
batch-upload /resources/data/<file_name>
```
Example
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-a /app/main user -n patientA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/patientA.priv
batch-upload /resources/data/example_a.csv
```

## Example. How to access shared data. Step by step.
1. Run help from prompt to get information about possible commands
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-admin /app/main user -h
```
2. Start and register 'admin' identity in the first place because Sawtooth blockchain requires admin user identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-admin /app/main user -n admin -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/admin.priv
```
3. Type in 'exit' in the prompt. To force exit use Ctrl + C
4. Start and register 'patientA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-a /app/main user -n patientA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/patientA.priv
```
5. Type in 'exit' in the prompt. To force exit use Ctrl + C
6. Start and register 'patientB' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-b /app/main user -n patientB -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/patientB.priv
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
12. Start 'patientA' identity
```
docker run -t -i --rm --network docker_default -v "$(pwd)"/resources/data:/resources/data docker_healthcare-system-client-patient-a /app/main user -n patientA -u rest-api-0:8008 -V tcp://validator-0:4004 -k /app/resources/keys/patientA.priv
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
docker rmi $(docker images | grep 'hyperledger' | awk '{print $3}') --force
docker rmi $(docker images | grep 'docker_' | awk '{print $3}') --force
docker volume rm $(docker volume ls | awk '{print $2}')
docker volume rm docker_poet-shared
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