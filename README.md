## Steps to follow

# STATEFULL IMPLEMENTATION

`go mod tidy
go build -o dhcp main.go`

server -> main.go

client -> client.go

`go run main.go`

`go run client.go`

**Docker Run**

`docker images`

`docker run -p 69:69/udp dhcp_example`

`docker run --entrypoint /app/bin/client dhcp_example`

**Docker repo**

`https://hub.docker.com/r/manishjha8909/dhcp_example`

pull from docker hub

`docker pull manishjha8909/dhcp_example`

# SERVERLESS IMPLEMENTATION

This project implements a serverless Dynamic Host Configuration Protocol (DHCP) service using:

- Knative Serving for event-driven scaling

- Redis for IP lease management

- Docker for containerization

- Minikube for local Kubernetes testing

directory structure

serverless/
├── server.go # Main server logic (HTTP-based DHCP-like)
├── dhcp_service.yaml # Knative Service definition
├── Dockerfile_server # Dockerfile for building dhcp-serverless
├── go.mod / go.sum # Go module files
└── README.md # This documentation

Prereqisites

Docker
Minikube with Knative installed
Helm (for Redis)
Knative Networking Layer (Kourier recommended)

Setup and deployment steps

1. Start minikube with Knative
   minikube start --memory=4096 --cpus=2

2. Install Redis using Helm
   helm install redis bitnami/redis --namespace redis --create-namespace

3. Extract and Duplicate Redis Secret to Default Namespace
   kubectl get secret redis -n redis -o yaml | \
    sed 's/namespace: redis/namespace: default/' > redis-secret.yaml

kubectl apply -f redis-secret.yaml

4. Build & Push the Docker Image
   docker build -t manishjha8909/dhcp-serverless:latest -f Dockerfile_server .
   docker push manishjha8909/dhcp-serverless:latest

5. Deploy the Knative Service
   kubectl apply -f dhcp_service.yaml

6. Force Knative to Redeploy (if necessary)
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

7. Test the API &
   kubectl run -it --rm curl-test --image=curlimages/curl --restart=Never -- \
    curl -v -X POST http://dhcp-service.default.svc.cluster.local \
    -H "Content-Type: application/json" \
    -d '{"type":"DISCOVER","mac":"11:22:33:44:55:66"}'

8. Check the Redis Lease
   REDIS_PASSWORD=$(kubectl get secret redis -n default -o jsonpath='{.data.redis-password}' | base64 -d)

kubectl exec -it redis-master-0 -n redis -- \
 redis-cli -a "$REDIS_PASSWORD" KEYS 'dhcp:lease:\*'

Output Example

{
"lease_time": 3600,
"offer": {
"expires": 1746214667,
"ip": "192.168.1.177",
"mac": "11:22:33:44:55:66"
}
}
