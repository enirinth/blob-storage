Interactive read clients
--------- 
## How to use
#### Always read from DC1
```sh
go run read_client.go
```
Type in (partition id, blob id) pair 
Will return a (content, size) pair
#### Read from nearest DC
```sh
go run read_from_nearest.go
```
The client will ping three DCs (IP address specified in cluster config), determine which one is the nearest, and all read requests will be sent to that one.   
Type in (partition id, blob id) pair after message finding the nearest one   
Will return a (content, size) pair
