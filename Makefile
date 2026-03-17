.PHONY: up down restart logs traffic clean

# Start all services
up:
	docker compose up -d --build

# Stop all services
down:
	docker compose down

# Stop and remove volumes
clean:
	docker compose down -v

# Restart everything
restart: down up

# View logs
logs:
	docker compose logs -f

# View specific service logs
logs-%:
	docker compose logs -f $*

# Generate traffic for 2 minutes
traffic:
	chmod +x scripts/generate_traffic.sh
	./scripts/generate_traffic.sh 120

# Check service health
status:
	@echo "── Services ──────────────────────"
	@docker compose ps
	@echo ""
	@echo "── App Health ────────────────────"
	@curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo "App not ready"
	@echo ""
	@echo "── Elasticsearch ─────────────────"
	@curl -s http://localhost:9200/_cluster/health | python3 -m json.tool 2>/dev/null || echo "ES not ready"
