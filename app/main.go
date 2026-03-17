package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ── Prometheus Metrics ───────────────────────────
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeConnections)
}

// ── Structured JSON Logger ───────────────────────
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Status    int    `json:"status,omitempty"`
	Duration  string `json:"duration,omitempty"`
	ClientIP  string `json:"client_ip,omitempty"`
	Error     string `json:"error,omitempty"`
}

var logstashConn net.Conn

func initLogstash() {
	addr := os.Getenv("LOGSTASH_HOST")
	if addr == "" {
		addr = "logstash:5000"
	}

	// Retry connection to Logstash (it may take time to start)
	for i := 0; i < 30; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			logstashConn = conn
			logJSON(LogEntry{Level: "info", Message: "Connected to Logstash"})
			return
		}
		log.Printf("Waiting for Logstash at %s (attempt %d/30)...", addr, i+1)
		time.Sleep(2 * time.Second)
	}
	log.Println("WARNING: Could not connect to Logstash. Logging to stdout only.")
}

func logJSON(entry LogEntry) {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	entry.Service = "sample-app"

	data, _ := json.Marshal(entry)

	// Always print to stdout
	fmt.Println(string(data))

	// Also send to Logstash if connected
	if logstashConn != nil {
		logstashConn.Write(append(data, '\n'))
	}
}

// ── Middleware: Logging + Metrics ─────────────────
func metricsMiddleware(next http.HandlerFunc, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Wrap ResponseWriter to capture status code
		wrapped := &statusWriter{ResponseWriter: w, status: 200}
		next(wrapped, r)

		duration := time.Since(start)
		status := fmt.Sprintf("%d", wrapped.status)

		// Record Prometheus metrics
		httpRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration.Seconds())

		// Structured log
		logJSON(LogEntry{
			Level:    "info",
			Message:  "request completed",
			Method:   r.Method,
			Path:     r.URL.Path,
			Status:   wrapped.status,
			Duration: duration.String(),
			ClientIP: r.RemoteAddr,
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ── Handlers ─────────────────────────────────────
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate variable latency
	delay := time.Duration(rand.Intn(200)) * time.Millisecond
	time.Sleep(delay)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from observability-lab!",
	})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate random errors for testing alerts/dashboards
	if rand.Float64() < 0.5 {
		logJSON(LogEntry{
			Level:   "error",
			Message: "simulated internal error",
			Error:   "something went wrong in processing",
		})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "no error this time",
	})
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate slow endpoint for latency monitoring
	delay := time.Duration(500+rand.Intn(2000)) * time.Millisecond
	time.Sleep(delay)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("responded after %v", delay),
	})
}

// ── Main ─────────────────────────────────────────
func main() {
	initLogstash()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", metricsMiddleware(healthHandler, "/health"))
	mux.HandleFunc("/hello", metricsMiddleware(helloHandler, "/hello"))
	mux.HandleFunc("/error", metricsMiddleware(errorHandler, "/error"))
	mux.HandleFunc("/slow", metricsMiddleware(slowHandler, "/slow"))
	mux.Handle("/metrics", promhttp.Handler())

	logJSON(LogEntry{
		Level:   "info",
		Message: "Server starting on :8080",
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
