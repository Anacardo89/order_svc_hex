# Tech stack  

## Services  

- Go 1.25.4  

### order_api  
```go.mod
require (
	github.com/caarlos0/env/v9 v9.0.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.13.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.65.0
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v3 v3.0.1
)
```  

### order_svc  
```go.mod
require (
	github.com/caarlos0/env/v9 v9.0.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.13.0
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.11.1
	github.com/testcontainers/testcontainers-go v0.40.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.40.0
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.7
	gopkg.in/yaml.v3 v3.0.1
)
```  

## Docker compose  

A simplified view of the services provisioned by docker compose:  
```yaml
services:
  # order_svc
  order_svc:
    build:
      context: ./order_svc
    container_name: order_svc
    depends_on:
      db:
        condition: service_healthy
      kafka:
        condition: service_healthy
      loki:
        condition: service_started
  # order_api
  order_api:
    build:
      context: ./order_api
    container_name: order_api
    depends_on:
      order_svc:
        condition: service_started
  # DB
  db:
    image: postgres:16
    container_name: order-svc-db
  # Kafka
  kafka:
    image: confluentinc/cp-kafka:7.6.1
    container_name: kafka
  # Loki
  loki:
    image: grafana/loki:2.9.0
    container_name: loki
  # Tempo
  tempo:
    image: grafana/tempo:2.4.0
    container_name: tempo
  #Grafana
  grafana:
    image: grafana/grafana:10.3.3
    container_name: grafana
    depends_on:
      - loki
      - tempo
```

