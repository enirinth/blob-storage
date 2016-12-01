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
sugo go get github.com/enirinth/blob-storage
```
## About go commands
I've created an alias as follows
```sh
alias sugo='sudo /bin/env GOPATH=\$HOME/workspace/gowork'
```
For security reasons, `sudo` does not honor some of the environmental variables, so you will get a `$GOPATH not set` error if you use `sudo go ....`    
i.e. **Always remember to use `sugo` before any go commands**

