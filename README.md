## Steps to follow

`go mod tidy
go build -o dhcp main.go`

server - main.go
client - client.go

go run main.go
go run client.go


Docker Run

docker images
docker run -p 69:69/udp dhcp_example
docker run --entrypoint /app/bin/client dhcp_example

# MTP_SNF

Master's Thesis Project
