
FROM golang:1.19 as builder


WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binaries
RUN go build -o bin/server main.go
RUN go build -o bin/client client.go

# lightweight runtime image
FROM debian:bullseye-slim

# Copy binaries from the builder
COPY --from=builder /app/bin /app/bin

EXPOSE 67/udp 68/udp 69/udp

# entry point 
ENTRYPOINT ["/app/bin/server"]
CMD []
