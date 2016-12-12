#!/bin/bash

### Write Client
numFile=$1
python zipf_0_2.py "$numFile"
go run write_client.go 0 input.txt

cd ../read-client
echo "first test"
go run central_manager_read_client.go 1

echo "second test"
go run central_manager_read_client.go 2

echo "third test"
go run central_manager_read_client.go 3


