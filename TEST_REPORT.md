## Process of bar chart creation
1. Run 'Test_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes.go' test
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client-admin go test -v test/1_5_10_users_share_data_test.go test/stats_test.go
```

2. Run 'Test_User_Register' test
```
docker run -t -i --rm --network docker_default docker_healthcare-system-client-admin go test -v test/1_5_10_users_register_data_test.go test/stats_test.go
```

3. All csv data reports should be generated in test/report directory with timestamp-name_of_test.csv pattern
4. Run chart generation scripts
```
go run cmd/charts/register_user/main.go
go run cmd/charts/share_data/main.go
```
5. All charts will generated in resources/charts directory in HTML format

## NOTE

* I could not performnce test more than 5 simultanious clients because blockchain network usually responds with 503 service unavailable. 
It happens due to laptop resources limits.

* Bar charts are not linear and have spikes. I assume it happens  because laptop processes can be busy by other processes. 
During evaulation it causes throughput to spike.

* User register X,Y axises correlate. It happens because client register function memory load is static and is not changing drastically.
