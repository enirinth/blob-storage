How to
--

## Set up AWS instances
```sh
bash aws-setup.sh
source ~/.bashrc
```
Run above after you login a **newly launched** AWS instances (EC2, free-tier, RHEL) i.e. Don't run `aws-setup.sh` more than once
## Update project code base
```sh
go get -u github.com/enirinth/blob-storage
```

## How to use
#### Server
Config `clusterconfig/constants.go` to set IP and ports for three DCs    
```sh
go run server/storage_server.go 1
```
Above command will start storage server for DC1, the running instance's IP has to match DC1's IP address specified in `clusterconfig/constants.go`    
#### Client
`Interactive-client` are for test/demo purposes, `client` is used for actual experiments. Go into folder for further details

