#!/bin/bash

### Write Client
numFile=$1
dcName=$2
python zipf_0_2.py "$numFile"
go run write_client.go 0 input.txt

cd ../read-client
echo "First test, 2 clients"
go run central_manager_read_client.go $2 2

echo "second tes, 5 clients"
go run central_manager_read_client.go $2 5

echo "third test, 10 clients"
go run central_manager_read_client.go $2 10

echo "third test, 20 clients"
go run central_manager_read_client.go $2 20

echo "third test, 40 clients"
go run central_manager_read_client.go $2 40

echo "third test, 80 clients"
go run central_manager_read_client.go $2 80
