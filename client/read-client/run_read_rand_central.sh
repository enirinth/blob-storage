#!/bin/bash

### Write Client
dcName=$1

echo "First test, 2 clients"
go run central_manager_read_client.go $1 2

echo "second tes, 5 clients"
go run central_manager_read_client.go $1 5

echo "third test, 10 clients"
go run central_manager_read_client.go $1 10

echo "third test, 20 clients"
go run central_manager_read_client.go $1 20

echo "third test, 30 clients"
go run central_manager_read_client.go $1 30

echo "third test, 40 clients"
go run central_manager_read_client.go $1 40

echo "third test, 50 clients"
go run central_manager_read_client.go $1 50

echo "third test, 80 clients"
go run central_manager_read_client.go $1 80
