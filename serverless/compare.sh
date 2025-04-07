#!/bin/bash
# Test serverfull
cd serverfull
time docker-compose run client

# Test serverless
cd ../serverless
kubectl apply -f knative/
time go run client/main.go