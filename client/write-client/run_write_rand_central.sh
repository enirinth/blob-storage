#!/bin/bash

### Write Client
numFile=$1
dcName=$2
python zipf_0_2.py "$numFile"
go run write_client.go 0 input.txt
