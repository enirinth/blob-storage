Interactive read clients
--------- 
## How to use
#### Write to a specific DC
```sh
go run write_client.go 1
```
Commandline argument could be either 1 2 or 3    
Type in (content, size) pair  (*Size is a fake number, not actual file size*)    
Will return a (partition id, blob id) pair
#### Write to nearest DC
```sh
go run write_to_nearest.go
```
The client will ping three DCs (IP address specified in cluster config), determine which one is the nearest, and all write requests will be sent to that one.   
Type in (content, size) pair after message finding the nearest one   
Will return a (partition id, blob id) pair
