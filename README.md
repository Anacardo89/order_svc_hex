# Order Service (Hexagonal/CQRS)
**Status:** Architectural PoC / Being worked on.

### Technical Highlights:
- **Architecture:** Hexagonal (Ports & Adapters) to decouple core logic from Kafka/gRPC.
- **CQRS:** Write-side via Kafka events; Read-side via gRPC.
- **Observability:** Distributed tracing propagated across service boundaries (Kafka headers + gRPC interceptors). 
- **Stack:** Go, gRPC, Kafka, Loki/Tempo.

### Todo
- Wire up Tempo
- Implement metrics/ wire up Mimir
- Make Graphana Dashboards
