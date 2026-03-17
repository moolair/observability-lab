# Observability Lab

A full-stack observability environment running locally with Docker Compose. Collects **logs**, **metrics**, and monitors a sample Go application through the ELK stack (Elasticsearch, Logstash, Kibana) and Prometheus + Grafana.

Built to demonstrate hands-on experience with production observability tooling.

## Architecture

```
┌─────────────┐       JSON logs        ┌───────────┐       ┌─────────────────┐
│  Sample App │ ────────────────────▶  │ Logstash  │ ────▶ │ Elasticsearch   │
│  (Go :8080) │                        └───────────┘       └────────┬────────┘
│             │                                                     │
│  /metrics   │ ◀── scrape ── ┌──────────────┐              ┌──────▼──────┐
│  /hello     │               │  Prometheus  │              │   Kibana    │
│  /error     │               │    (:9090)   │              │   (:5601)   │
│  /slow      │               └──────┬───────┘              └─────────────┘
│  /health    │                      │
└─────────────┘               ┌──────▼───────┐
                              │   Grafana    │
                              │   (:3000)    │
                              └──────────────┘
```

## What's Inside

| Component       | Purpose                              | Port  |
|----------------|--------------------------------------|-------|
| **Sample App**  | Go HTTP server with structured JSON logging + Prometheus metrics | 8080 |
| **Prometheus**  | Metrics collection and time-series storage | 9090 |
| **Grafana**     | Metrics visualization and dashboards | 3000 |
| **Elasticsearch** | Log storage and full-text search   | 9200 |
| **Logstash**    | Log ingestion pipeline (JSON → ES)   | 5000 |
| **Kibana**      | Log exploration and visualization    | 5601 |

## Quick Start

### Prerequisites
- Docker & Docker Compose
- 8GB+ RAM (ELK is memory-hungry)
- `curl` for load testing

### Run

```bash
# Clone and start everything
git clone https://github.com/moolair/observability-lab.git
cd observability-lab

# Start all services
docker compose up -d --build

# Wait ~30s for Elasticsearch and Logstash to initialize
# Then generate some traffic
chmod +x scripts/generate_traffic.sh
./scripts/generate_traffic.sh 120
```

### Access Dashboards

| Service     | URL                        | Credentials   |
|-------------|----------------------------|---------------|
| Grafana     | http://localhost:3000       | admin / admin |
| Kibana      | http://localhost:5601       | —             |
| Prometheus  | http://localhost:9090       | —             |
| Sample App  | http://localhost:8080/hello | —             |

### Stop

```bash
docker compose down -v   # -v removes data volumes
```

## Sample App Endpoints

| Endpoint   | Behavior                                          |
|-----------|---------------------------------------------------|
| `/hello`  | Returns greeting, random latency (0-200ms)         |
| `/health` | Health check, always 200                           |
| `/error`  | 50% chance of 500 error (for testing alerts)       |
| `/slow`   | Slow response (500ms-2.5s) for latency monitoring  |
| `/metrics`| Prometheus metrics endpoint                        |

### Metrics Exposed

- `http_requests_total` — counter by method, endpoint, status
- `http_request_duration_seconds` — histogram of request latency
- `active_connections` — gauge of in-flight requests

## Exploring the Data

### Kibana (Logs)
1. Go to http://localhost:5601
2. Create a Data View: `app-logs-*` with `@timestamp`
3. Open **Discover** → filter by `level: error` to see simulated errors
4. Try Lucene query: `status:500 AND path:"/error"`

### Prometheus (Metrics)
1. Go to http://localhost:9090
2. Try these PromQL queries:
   - `rate(http_requests_total[1m])` — request rate per second
   - `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` — p95 latency
   - `http_requests_total{status="500"}` — total 5xx errors

### Grafana (Dashboards)
1. Go to http://localhost:3000 (admin/admin)
2. Prometheus and Elasticsearch are auto-provisioned as datasources
3. Create a dashboard with panels for request rate, error rate, and latency

## Project Structure

```
observability-lab/
├── app/
│   ├── main.go              # Go app with metrics + structured logging
│   ├── go.mod
│   └── Dockerfile
├── prometheus/
│   └── prometheus.yml       # Scrape config
├── logstash/
│   └── pipeline/
│       └── logstash.conf    # Log ingestion pipeline
├── grafana/
│   └── provisioning/
│       └── datasources/
│           └── datasources.yml  # Auto-configured data sources
├── scripts/
│   └── generate_traffic.sh  # Load generator
├── docker-compose.yml
└── README.md
```

## Tech Stack

- **Go** — sample application with Prometheus client library
- **Prometheus** — metrics scraping and storage
- **Grafana** — dashboards and alerting
- **Elasticsearch** — log indexing and search
- **Logstash** — log pipeline with JSON parsing and enrichment
- **Kibana** — log exploration with Lucene queries
- **Docker Compose** — single-command orchestration

## Roadmap

- [ ] Add OpenTelemetry Collector for distributed tracing
- [ ] Coralogix integration via OpenTelemetry exporter
- [ ] Kubernetes deployment with Helm charts
- [ ] Grafana dashboard JSON export (pre-built dashboards)
- [ ] Alertmanager rules for error rate thresholds

## License

MIT
# observability-lab
