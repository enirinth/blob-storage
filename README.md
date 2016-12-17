# README
--

## Set up instances
```sh
bash aws-setup.sh
source ~/.bashrc
```
It's best to run above after you login a **newly launched** AWS instances (EC2, free-tier, RHEL) i.e. Don't run `aws-setup.sh` more than once. This should work for most linux distro with YUM
## Update project code base
```sh
go get -u github.com/enirinth/blob-storage
```

## project directory
Enter project directory
```sh
cd582
```

## Config
We can update number of DCs, IP, port and many other settings in `clusterconfig/constants.go`. (It is already 0.0.0.0, you are safe to test this locally)   

## Centralized design
#### Server
```sh
cd server                   # enter server folder
go run central_manager.go   # run central manager
go run central_dc.go 1      # run data center server
go run central_dc.go 2
go run central_dc.go 3
```
Now our servers are ready, waiting for read/write requests.

#### Client
```sh
cd client/write-client             # enter write client folder to create blobs
sh run_central_write_blob.sh 500   # create specific number of files which follows the zipf distribution
cd client/read-client              # enter read client folder to generate read requests
sh central_manager_read_test.sh    # start read requests
```
After those operations, we have generate 500 blobs and test read requests for centralized manager.

## De-centralized Design
#### Server 
```sh
go run server/storage_server.go 1
```
This will start storage server on DC 1     
Similar to the other DCs, remember to change the last parameter. 2 for DC2, 3 for DC3.

#### Client
Interactive clients are ligth-wighted for communication with servers, like write one blob file or read specific blob.
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
