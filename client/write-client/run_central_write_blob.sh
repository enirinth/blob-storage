#!/bin/bash

### Write Client
numFile=$1
cd ../../data/
python zipf_0_2.py "$numFile"

cd ../client/write-client/
go run central_manager_write_blob.go input.txt
