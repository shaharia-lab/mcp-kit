#!/bin/bash
# Setup script for the monitoring stack

# Create directories
mkdir -p monitoring/prometheus

# Create Prometheus configuration file
cat > monitoring/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node_exporter'
    static_configs:
      - targets: ['node_exporter:9100']

  - job_name: 'mcp-server'
    metrics_path: /metrics
    static_configs:
      - targets: ['mcp-server:8080']

  - job_name: 'mcp-client'
    metrics_path: /metrics
    static_configs:
      - targets: ['mcp-client:8081']
EOF

echo "Monitoring setup completed. Now you can start the stack with 'docker-compose up -d'"