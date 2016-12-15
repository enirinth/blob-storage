#!/bin/bash

### Write Client
dc=0
num=6

echo "2 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 2
done

echo
echo
echo "5 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 5
done

echo
echo
echo "10 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 10
done

echo
echo
echo "20 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 20
done

echo
echo
echo "30 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 30
done

echo
echo
echo "40 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 40
done

echo
echo
echo "50 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 50
done

echo
echo
echo "60 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 60
done

echo
echo
echo "80 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 80
done

echo
echo
echo "100 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 100
done

echo
echo
echo "120 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 120
done

echo
echo
echo "150 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 150
done

echo
echo
echo "200 clients"
for i in $(seq 1 $num); do
    go run central_manager_read_client.go $dc 200
done
