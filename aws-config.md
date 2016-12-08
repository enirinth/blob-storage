Aws instances info
----------
## DC 1 -  Ireland
#### SSH using key-pair
```sh
ssh -i "keypair-ireland.pem" ec2-user@ec2-52-209-171-220.eu-west-1.compute.amazonaws.com
```
#### SSH after registering pub key 
```sh
ssh ec2-user@ec2-52-209-171-220.eu-west-1.compute.amazonaws.com
```
#### Public IP
```sh
52.209.171.220
```

## DC 2 - North Va
#### SSH using key-pair
```sh
ssh -i "keypair-northva.pem" ec2-user@ec2-54-221-133-142.compute-1.amazonaws.com  
```
#### SSH after registering pub key                   
```sh
ssh ec2-user@ec2-54-221-133-142.compute-1.amazonaws.com 
```
#### Public IP                                             
```sh
54.221.133.142
```

## DC 3 -  North Cal
#### SSH using key-pair
```sh
ssh -i "keypair-northcal.pem" ec2-user@ec2-54-153-39-155.us-west-1.compute.amazonaws.com 
```
#### SSH after registering pub key              
```sh
ssh ec2-user@ec2-54-153-39-155.us-west-1.compute.amazonaws.com
```
#### Public IP
```sh
54.153.39.155
```
