#!/bin/bash
set -e
set -x

echo "Starting Promster..."
promster \
    -log-level=$LOG_LEVEL \
    --scrape-shard-enable=$SCRAPE_SHARD_ENABLE \
    --scrape-etcd-url=$SCRAPE_ETCD_URL \
    --scrape-etcd-path=$SCRAPE_ETCD_PATH \
    --registry-etcd-url=$REGISTRY_ETCD_URL \
    --registry-etcd-base=$REGISTRY_ETCD_BASE \
    --registry-service-name=$REGISTRY_SERVICE \
    --registry-node-ttl=$REGISTRY_TTL&

echo "Starting Prometheus..."
prometheus --config.file=/prometheus.yml --web.enable-lifecycle --storage.tsdb.retention=$RETENTION_TIME



