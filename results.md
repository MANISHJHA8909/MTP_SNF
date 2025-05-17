---
layout: page
title: Results & Evaluation
permalink: /results/
---

# Results & Evaluation

This section presents the performance comparison between **Stateful** and **Serverless (Cold Start)** DHCP implementations.

## Latency Comparison Graph

![Latency Graph](/assets/images/Result_Graph.png)

**Figure 6.1**: Latency over 20 attempts comparing Stateful and Serverless (Cold Start) DHCP implementations.

## Analysis

- **Stateful Implementation**:
  - First attempt latency: ~17.61s (due to binary/resource loading).
  - Subsequent attempts: Stabilized around **1.6–1.7s**.
- **Serverless (Cold Start)**:
  - Every attempt triggered a cold start due to scaling to 0.
  - Latency observed: **5–8 seconds** consistently.
  - Due to Knative provisioning and container spin-up overhead.

### Conclusion

Despite the cold start penalty, serverless provides better **resource efficiency** and **auto-scaling**, making it suitable for bursty workloads. Stateful systems offer **lower latency** for consistent workloads but require persistent resource allocation.
