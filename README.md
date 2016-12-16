How to
--

## Set up instances
```sh
bash aws-setup.sh
source ~/.bashrc
```
It's best to run above after you login a **newly launched** AWS instances (EC2, free-tier, RHEL) i.e. Don't run `aws-setup.sh` more than once    
This should work for most linux distro with YUM
## Update project code base
```sh
go get -u github.com/enirinth/blob-storage
```

## Demo
Enter project directory
```sh
cd582
```
#### Server
Config `clusterconfig/constants.go` to set IP and ports for three DCs. (It is already 0.0.0.0, you are safe to test this locally)    
```sh
go run server/storage_server.go 1
```
This will start storage server on DC 1     
Start DC2 and DC3 on different terminal sessions   
#### Client
Open another terminal session and run
```sh
go run interactive-client/write-client/write_client.go 1
```
This will start a write client writing to DC1, after initialization, type in a string (file content) and a number (file size), separted by space, e.g.
```sh
test 1
```
It will return a (`<id1>, <id2>`) pair   
Check DC1's terminal session, it will display the updated storage table   
Close the write client, and run    
```sh
go run interactive-client/read-client/read_from_nearest.go
```
This will start a read client, ping all three DCs, decide which one is the nearest and send request     
Type in the pair you got from write request  
```sh
<id1, id2>
```
It will return `test 1`    
Read the same `<id1, id2>` more than 5 times, and check the servers, you will see the partition containing this file gets populated to all DCs   
