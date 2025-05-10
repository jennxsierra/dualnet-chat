#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <port>"
    exit 1
fi

PORT=$1

PID=$(sudo lsof -i :$PORT | awk 'NR==2 {print $2}')

if [ -z "$PID" ]; then
    echo "No process found listening on port $PORT."
    exit 1
fi

echo "Found process with PID $PID on port $PORT."
echo ""
pmap $PID | tail -n 1
