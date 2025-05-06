#!/bin/bash

# Variables
DOCKER_USERNAME="manishjha8909"
SERVERLESS_IMAGE="dhcp-serverless"
SERVERLESS_TAG="latest"
STATEFUL_IMAGE="dhcp_example"
STATEFUL_TAG="latest"
REGISTRY="docker.io"

# Deploy Kubernetes namespace (optional)
NAMESPACE="default"

# DockerHub Credentials (if needed, uncomment and use)
# echo "Enter DockerHub username:"
# read DOCKER_USERNAME
# echo "Enter DockerHub password:"
# read -s DOCKER_PASSWORD
# docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD

# 1. Build Docker images for both serverless and stateful

echo "Building Docker image for serverless DHCP..."
docker build -t $DOCKER_USERNAME/$SERVERLESS_IMAGE:$SERVERLESS_TAG -f serverless/Dockerfile_server .
if [ $? -ne 0 ]; then
    echo "Failed to build serverless DHCP Docker image"
    exit 1
fi

echo "Building Docker image for stateful DHCP..."
docker build -t $DOCKER_USERNAME/$STATEFUL_IMAGE:$STATEFUL_TAG -f stateful/Dockerfile .
if [ $? -ne 0 ]; then
    echo "Failed to build stateful DHCP Docker image"
    exit 1
fi

# 2. Push Docker images to Docker Hub
echo "Pushing serverless Docker image to Docker Hub..."
docker push $DOCKER_USERNAME/$SERVERLESS_IMAGE:$SERVERLESS_TAG
if [ $? -ne 0 ]; then
    echo "Failed to push serverless DHCP Docker image"
    exit 1
fi

echo "Pushing stateful Docker image to Docker Hub..."
docker push $DOCKER_USERNAME/$STATEFUL_IMAGE:$STATEFUL_TAG
if [ $? -ne 0 ]; then
    echo "Failed to push stateful DHCP Docker image"
    exit 1
fi

# 3. Deploy the serverless service to Knative

echo "Deploying Knative service..."
kubectl apply -f serverless/dhcp_service.yaml
if [ $? -ne 0 ]; then
    echo "Failed to deploy Knative service"
    exit 1
fi

# 4. Redeploy if needed (force update)
echo "Forcing Knative redeployment (if necessary)..."
kubectl patch ksvc dhcp-service --type=merge -p '{
   "spec": {
     "template": {
       "metadata": {
         "annotations": {
           "autoscaling.knative.dev/minScale": "1",
           "dhcp/update-time": "'$(date +%s)'"
         }
       }
     }
   }
 }'

# 5. Verify Deployment

echo "Checking Knative service status..."
kubectl get ksvc dhcp-service
if [ $? -ne 0 ]; then
    echo "Failed to get Knative service status"
    exit 1
fi

echo "Deployment successful! You can now test the API and check Redis as per the README instructions."

