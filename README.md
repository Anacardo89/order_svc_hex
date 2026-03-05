# Order Service (Hexagonal/CQRS)
**Status:** Architectural PoC / Being worked on.

[Gist with project analysis](https://gist.github.com/Anacardo89/44471e8ef57e71b11e1c184a6dcfcdb1)
![Tracing](trace.jpg)

### Technical Highlights:
- **Architecture:** Hexagonal (Ports & Adapters) to decouple core logic from Kafka/gRPC.
- **CQRS:** Write-side via Kafka events; Read-side via gRPC.
- **Observability:** Distributed tracing propagated across service boundaries (Kafka headers + gRPC interceptors). 
- **Stack:** Go, gRPC, Kafka, Loki/Tempo.

### Todo
- Implement metrics/ wire up Mimir
- Make Graphana Dashboards
