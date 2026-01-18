#!/bin/bash
set -e

echo "Waiting for Kafka to be ready..."
# Wait until Kafka responds
until kafka-topics --bootstrap-server localhost:9092 --list; do
  echo "Kafka not ready, sleeping 2s..."
  sleep 2
done

echo "Kafka ready! Creating topics..."

TOPICS=(
  "orders.created"
  "orders.statusUpdated"
  "orders.created.dlq"
  "orders.statusUpdated.dlq"
  "orders.transport.dlq"
)

for topic in "${TOPICS[@]}"; do
  if ! kafka-topics --bootstrap-server localhost:9092 --list | grep -q "^$topic$"; then
    kafka-topics --bootstrap-server localhost:9092 \
                 --create \
                 --topic "$topic" \
                 --partitions 1 \
                 --replication-factor 1
  fi
  sleep 1
done

echo "All topics created!"
