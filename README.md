How to use
--
## What is this project?
A mock storage server that stores BLOBs(Binary Large OBjects) in geographically different data centers in a fashion that:
- Selectively replicate BLOBs based on their popularity
- Maintaining a auto-configurable global state machine
- Introduced a novel distributed consensus protocol      

More details can be found in `selective-data-replication.pdf`


## Set up instances
```sh
bash aws-setup.sh
source ~/.bashrc
```
It's best to run above after you login a **newly launched** AWS instances (EC2, free-tier, RHEL) i.e. Don't run `aws-setup.sh` more than once. This should work for most linux distro with YUM

```sh
go get -u github.com/enirinth/blob-storage    # Update project code base
```

An alias is already created to enter project root directory   
```sh
cd582    
```

## Configuration
You may configure number of DCs, IP, port and other settings in `clusterconfig/constants.go`.
(All DC IPs are 0.0.0.0 by default, you are safe to test this locally)   

## Demo for De-centralized Server
Go to project root directory
```sh
cd582
```
#### Server 
```sh
go run server/storage_server.go 1
```
This will start DC server 1.   
Similarly start server 2 and 3 in different terminal sessions.    

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

## Demo for Centralized Server
#### Server
```sh
cd server                   # enter server folder
go run central_manager.go   # run central manager
go run central_dc.go 1      # run DC storage server
go run central_dc.go 2     
go run central_dc.go 3
```
Start DC servers and central manager.
Remember to run central manager and three DC servers in different terminal sessions.

#### Client
```sh
cd client/write-client             # enter write client folder to create blobs
sh run_central_write_blob.sh 500   # create specific number of files which follows the zipf distribution
cd client/read-client              # enter read client folder to generate read requests
sh central_manager_read_test.sh    # start read requests
```
500 blobs will be generated and latency results will be printed out
