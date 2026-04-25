# Project Structure

```bash
.
├── configs
│   ├── grafana
│   │   └── provisioning
│   │       └── datasources
│   │           └── ds.yaml
│   ├── loki-config.yaml
│   ├── prometheus-config.yaml
│   └── tempo-config.yaml
├── contracts
│   └── orders
│       └── order.proto
├── deployments
│   ├── docker
│   │   ├── docker-compose.yaml
│   │   └── sample.env
│   └── k8s
│       ├── kind-config.yaml
│       └── Makefile
├── docs
│   ├── directory_tree.md
│   ├── project_blueprint.md
│   └── tech_stack.md
├── order_api
│   ├── cmd
│   │   └── main
│   │       ├── init.go
│   │       └── main.go
│   ├── config
│   │   ├── config.go
│   │   ├── config.yaml
│   │   └── models.go
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── adapters
│   │   │   ├── in
│   │   │   │   └── http
│   │   │   │       └── rest
│   │   │   │           └── orderorchestrator
│   │   │   │               ├── errors.go
│   │   │   │               ├── handlers.go
│   │   │   │               ├── metrics.go
│   │   │   │               ├── middleware.go
│   │   │   │               ├── router.go
│   │   │   │               ├── server.go
│   │   │   │               └── service.go
│   │   │   ├── infra
│   │   │   │   └── log
│   │   │   │       └── loki
│   │   │   │           └── logger
│   │   │   │               ├── helpers.go
│   │   │   │               ├── logger.go
│   │   │   │               └── logger_handler.go
│   │   │   └── out
│   │   │       ├── messaging
│   │   │       │   └── kafka
│   │   │       │       └── orderwriter
│   │   │       │           ├── client.go
│   │   │       │           ├── handlers.go
│   │   │       │           ├── metrics.go
│   │   │       │           ├── models.go
│   │   │       │           └── producer.go
│   │   │       └── rpc
│   │   │           └── grpc
│   │   │               └── orderreader
│   │   │                   ├── client.go
│   │   │                   ├── handlers.go
│   │   │                   ├── interceptors.go
│   │   │                   └── models.go
│   │   ├── core
│   │   │   ├── models.go
│   │   │   └── models_status.go
│   │   └── ports
│   │       ├── logger.go
│   │       ├── order_orchestrator.go
│   │       ├── order_reader.go
│   │       └── order_writer.go
│   ├── pkg
│   │   ├── events
│   │   │   └── connection.go
│   │   ├── observability
│   │   │   ├── metrics.go
│   │   │   └── tracing.go
│   │   ├── ptr
│   │   │   └── ptr.go
│   │   └── validator
│   │       ├── validator.go
│   │       ├── validator_helpers.go
│   │       └── validator_test.go
│   └── proto
│       └── orderpb
│           ├── order_grpc.pb.go
│           └── order.pb.go
├── order_svc
│   ├── cmd
│   │   └── main
│   │       ├── init.go
│   │       └── main.go
│   ├── config
│   │   ├── config.go
│   │   ├── config.yaml
│   │   └── models.go
│   ├── db
│   │   ├── migrations
│   │   │   ├── 00001_function_set_updated.down.sql
│   │   │   ├── 00001_function_set_updated.up.sql
│   │   │   ├── 00002_enum_order_status.down.sql
│   │   │   ├── 00002_enum_order_status.up.sql
│   │   │   ├── 00003_table_orders.down.sql
│   │   │   ├── 00003_table_orders.up.sql
│   │   │   ├── 0004_trigger_updated_orders.down.sql
│   │   │   └── 0004_trigger_updated_orders.up.sql
│   │   └── seeds
│   │       └── orders_test.sql
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── adapters
│   │   │   ├── in
│   │   │   │   ├── messaging
│   │   │   │   │   └── kafka
│   │   │   │   │       └── orderconsumer
│   │   │   │   │           ├── client.go
│   │   │   │   │           ├── consumer.go
│   │   │   │   │           ├── handlers.go
│   │   │   │   │           ├── metrics.go
│   │   │   │   │           ├── models.go
│   │   │   │   │           └── routing.go
│   │   │   │   └── rpc
│   │   │   │       └── grpc
│   │   │   │           └── orderserver
│   │   │   │               ├── interceptors.go
│   │   │   │               ├── metrics.go
│   │   │   │               ├── models.go
│   │   │   │               ├── server.go
│   │   │   │               ├── server_handlers.go
│   │   │   │               ├── service.go
│   │   │   │               └── service_handlers.go
│   │   │   ├── infra
│   │   │   │   └── log
│   │   │   │       └── loki
│   │   │   │           └── logger
│   │   │   │               ├── logger.go
│   │   │   │               └── logger_handler.go
│   │   │   └── out
│   │   │       ├── messaging
│   │   │       │   └── kafka
│   │   │       │       └── orderdlq
│   │   │       │           ├── client.go
│   │   │       │           ├── handlers.go
│   │   │       │           ├── models.go
│   │   │       │           └── producer.go
│   │   │       └── store
│   │   │           └── pgx
│   │   │               └── orderrepo
│   │   │                   ├── error.go
│   │   │                   ├── main_test.go
│   │   │                   ├── models.go
│   │   │                   ├── orders.go
│   │   │                   ├── orders_test.go
│   │   │                   ├── repo.go
│   │   │                   └── test_utils.go
│   │   ├── core
│   │   │   └── models.go
│   │   └── ports
│   │       ├── logger.go
│   │       ├── order_consumer.go
│   │       ├── order_dlq.go
│   │       ├── order_repo.go
│   │       └── order_server.go
│   ├── pkg
│   │   ├── db
│   │   │   ├── connection.go
│   │   │   └── utils.go
│   │   ├── events
│   │   │   └── connection.go
│   │   ├── observability
│   │   │   ├── metrics.go
│   │   │   └── tracing.go
│   │   ├── ptr
│   │   │   └── ptr.go
│   │   └── testutils
│   │       ├── fs.go
│   │       └── postgres.go
│   └── proto
│       └── orderpb
│           ├── order_grpc.pb.go
│           └── order.pb.go
├── README.md
└── trace.jpg
```