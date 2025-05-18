---
layout: default
title: Serverless Network Function - DHCP
---

# Serverless Network Function (SNF) Project

This project demonstrates a comparison between **Stateful DHCP** and **Serverless DHCP** implementations using Docker, Redis, Kubernetes, and Knative.

---

## 🔧 Stateful Implementation

- Written in Go.
- Implements a basic DHCP client and server.
- Uses UDP sockets and a fixed lease database.
- Supports Docker-based execution.

**To Run:**

```bash
go mod tidy
go build -o dhcp main.go
go run main.go  # Server
go run client.go  # Client
```
