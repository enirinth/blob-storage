#!/bin/bash

### Write Client
numFile=$1
python zipf_0_2.py "$numFile"
go run write_client_rand.go input.txt


