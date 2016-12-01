How to
--

## Set up AWS instances
```sh
bash aws-setup.sh
source ~/.bashrc
```
## Update project code base
```sh
sugo go get github.com/enirinth/blob-storage
```
## About go commands
I've created an alias as follows
```sh
alias sugo='sudo /bin/env GOPATH=\$HOME/workspace/gowork'
```
For security reasons, `sudo` does not honor some of the environmental variables, so you will get a `$GOPATH not set` error if you use `sudo go ....`    
So **Always remeber to use `sugo` before any go commands**
