---
layout: default
title: Serverless Network Function - DHCP
---

<head>
  <link rel="stylesheet" href="{{ site.baseurl }}/assets/style.css">
</head>

<div style="text-align: center;">
  <h1>🚀 Serverless Network Function (SNF) Project</h1>
  <p><strong>Comparison between Stateful and Serverless DHCP Implementations</strong></p>
  <a href="#stateful" class="button">Jump to Stateful</a>
  <a href="#serverless" class="button">Jump to Serverless</a>
</div>

---

## 📌 Overview

This project demonstrates a side-by-side comparison between **Stateful DHCP** and **Serverless DHCP** implementations using:

- ⚙️ Docker
- 🧠 Redis
- ☸️ Kubernetes
- ☁️ Knative

We analyze latency, resource usage, and cold start time.

---

## 📊 Architecture Diagram

<img src="{{ site.baseurl }}/assets/images/architecture.png" alt="Architecture Diagram" />

---

## 🔧 <a name="stateful"></a>Stateful Implementation

<section>
<h3>📂 Features:</h3>
<ul>
  <li>Written in <strong>Go</strong></li>
  <li>Basic DHCP server + client</li>
  <li>Uses <strong>UDP sockets</strong> and a static lease database</li>
  <li>Deployed with Docker</li>
</ul>

<h3>▶️ Run Commands:</h3>

```bash
go mod tidy
go build -o dhcp main.go
go run main.go     # Start Server
go run client.go   # Start Client
```
