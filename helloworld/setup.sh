#!/bin/bash

for i in {1..9}
do
    port=900$i
    echo "Starting instance on port $port"
    go run . --server_address :$port &
done

wait