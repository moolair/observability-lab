#!/bin/bash
# ─────────────────────────────────────────────────
# generate_traffic.sh
# Sends requests to all endpoints to populate
# Prometheus metrics and Elasticsearch logs.
# Usage: ./scripts/generate_traffic.sh [duration_seconds]
# ─────────────────────────────────────────────────

DURATION=${1:-60}
BASE_URL="http://localhost:8080"
ENDPOINTS=("/hello" "/health" "/error" "/slow")

echo "🚀 Generating traffic for ${DURATION}s..."
echo "   Target: ${BASE_URL}"
echo "   Endpoints: ${ENDPOINTS[*]}"
echo ""

END_TIME=$((SECONDS + DURATION))
COUNT=0

while [ $SECONDS -lt $END_TIME ]; do
  # Pick a random endpoint
  ENDPOINT=${ENDPOINTS[$RANDOM % ${#ENDPOINTS[@]}]}
  
  # Send request (suppress output, show errors)
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${ENDPOINT}")
  COUNT=$((COUNT + 1))
  
  echo "[$(date '+%H:%M:%S')] ${ENDPOINT} → ${HTTP_CODE}"
  
  # Random delay between 0.1s and 1s
  sleep "0.$(( RANDOM % 9 + 1 ))"
done

echo ""
echo "✅ Done! Sent ${COUNT} requests in ${DURATION}s"
echo ""
echo "📊 Check your dashboards:"
echo "   Grafana:     http://localhost:3000  (admin/admin)"
echo "   Kibana:      http://localhost:5601"
echo "   Prometheus:  http://localhost:9090"
