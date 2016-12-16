#!/bin/bash

### Write Client
numFile=$1

cd ../../data/
python zipf_0_2.py "$numFile"

cd ../client/write-client/
go run write_client_rand.go input.txt
