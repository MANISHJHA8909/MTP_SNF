## Steps to follow

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


# MTP_SNF

Master's Thesis Project
curl -X POST http://dhcp-service.default.knative.local -d '{"type":"DISCOVER","mac":"11:22:33:44:55:66"}'

http://dhcp-service.default.knative.local
