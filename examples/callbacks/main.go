package main

import (
	"os"
	"runtime"
	"time"

	"github.com/bresrch/sawmill"
)

func main() {
	// Base logger for demonstration
	baseLogger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.WithPrettyPrint(true),
	))

	// === Basic Callback - Add Runtime Information ===

	// Callback that adds server information to every log
	serverInfoLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		hostname, _ := os.Hostname()
		record.WithDot("server.hostname", hostname)
		record.WithDot("server.pid", os.Getpid())
		record.WithDot("server.runtime.goroutines", runtime.NumGoroutine())

		// Memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		record.WithDot("server.memory.alloc_mb", float64(m.Alloc)/1024/1024)
		record.WithDot("server.memory.sys_mb", float64(m.Sys)/1024/1024)

		return record
	})

	serverInfoLogger.Info("User login", "user_id", 123, "username", "alice")

	// === Request Tracing Callback ===

	// Simulate request tracing with unique IDs
	requestID := "req-abc-123-def-456"
	correlationID := "corr-xyz-789"

	tracingLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		record.WithDot("trace.request_id", requestID)
		record.WithDot("trace.correlation_id", correlationID)
		record.WithDot("trace.timestamp", time.Now().UnixNano())
		record.WithDot("trace.service", "user-service")
		record.WithDot("trace.version", "v1.2.3")
		return record
	})

	tracingLogger.Info("Processing API request", "endpoint", "/api/users", "method", "GET")
	tracingLogger.Info("Database query executed", "table", "users", "duration_ms", 25)
	tracingLogger.Info("Response sent", "status_code", 200, "response_size_bytes", 1024)

	// === Performance Monitoring Callback ===

	// Callback that adds performance metrics
	perfLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		// CPU and memory info
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		record.WithDot("perf.timestamp", time.Now().Unix())
		record.WithDot("perf.goroutines", runtime.NumGoroutine())
		record.WithDot("perf.memory.heap_alloc_mb", float64(m.HeapAlloc)/1024/1024)
		record.WithDot("perf.memory.heap_sys_mb", float64(m.HeapSys)/1024/1024)
		record.WithDot("perf.memory.gc_runs", m.NumGC)
		record.WithDot("perf.cpu.cores", runtime.NumCPU())

		return record
	})

	perfLogger.Info("Heavy computation started", "algorithm", "matrix_multiplication")
	perfLogger.Info("Heavy computation completed", "duration_ms", 1500, "result_size", 10000)

	// === Conditional Callback - Error Context ===

	// Callback that adds extra context only for error logs
	errorContextLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		if record.Level >= sawmill.LevelError {
			// Add debugging context for errors
			record.WithDot("debug.stack_size", runtime.NumGoroutine())
			record.WithDot("debug.caller_info", "main.go:123")
			record.WithDot("debug.build_time", "2025-05-31T10:00:00Z")
			record.WithDot("debug.git_commit", "abc123def456")

			// Environment info
			record.WithDot("env.go_version", runtime.Version())
			record.WithDot("env.os", runtime.GOOS)
			record.WithDot("env.arch", runtime.GOARCH)
		}
		return record
	})

	errorContextLogger.Info("Normal operation")                                           // No extra context
	errorContextLogger.Error("Database connection failed", "error", "connection timeout") // Extra context added

	// === User Context Callback ===

	// Simulate user session context
	userSession := map[string]interface{}{
		"user_id":    12345,
		"username":   "alice_smith",
		"role":       "admin",
		"session_id": "sess-abc-123",
		"ip_address": "192.168.1.100",
		"user_agent": "Mozilla/5.0 Chrome/91.0",
	}

	userContextLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		for key, value := range userSession {
			record.WithDot("user."+key, value)
		}
		record.WithDot("user.session.start_time", time.Now().Add(-2*time.Hour))
		record.WithDot("user.session.last_activity", time.Now().Add(-5*time.Minute))
		return record
	})

	userContextLogger.Info("User action", "action", "view_profile")
	userContextLogger.Info("User action", "action", "update_settings", "changes", []string{"theme", "notifications"})

	// === Chained Callbacks ===

	// Multiple callbacks can be chained together
	chainedLogger := baseLogger.
		WithCallback(func(record *sawmill.Record) *sawmill.Record {
			// First callback: Add request context
			record.WithDot("request.id", "req-chain-001")
			record.WithDot("request.start_time", time.Now())
			return record
		}).
		WithCallback(func(record *sawmill.Record) *sawmill.Record {
			// Second callback: Add service context
			record.WithDot("service.name", "payment-processor")
			record.WithDot("service.instance", "instance-05")
			return record
		}).
		WithCallback(func(record *sawmill.Record) *sawmill.Record {
			// Third callback: Add environment context
			record.WithDot("env.region", "us-east-1")
			record.WithDot("env.availability_zone", "us-east-1a")
			return record
		})

	chainedLogger.Info("Payment processed", "amount", 99.99, "currency", "USD")

	// === Dynamic Callback - Load Balancer Info ===

	// Callback that adds different information based on conditions
	loadBalancerLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		currentTime := time.Now()

		// Simulate load balancer selection based on current second
		lbInstance := "lb-01"
		if currentTime.Second()%2 == 0 {
			lbInstance = "lb-02"
		}

		record.WithDot("lb.instance", lbInstance)
		record.WithDot("lb.region", "us-west-2")
		record.WithDot("lb.health_check", "healthy")
		record.WithDot("lb.active_connections", 150+currentTime.Second())
		record.WithDot("lb.requests_per_second", 45+currentTime.Second()%10)

		return record
	})

	loadBalancerLogger.Info("Request routed", "backend", "api-server-03")
	time.Sleep(time.Second) // Ensure different timestamp
	loadBalancerLogger.Info("Request routed", "backend", "api-server-01")

	// === Callback with Complex Nested Data ===

	// Callback that builds complex nested structures
	complexLogger := baseLogger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
		// Add comprehensive application context
		record.WithDot("app.name", "e-commerce-api")
		record.WithDot("app.version", "2.1.4")
		record.WithDot("app.build.number", "1234")
		record.WithDot("app.build.timestamp", "2025-05-31T08:00:00Z")

		// Infrastructure context
		record.WithDot("infra.cloud.provider", "aws")
		record.WithDot("infra.cloud.region", "us-east-1")
		record.WithDot("infra.kubernetes.cluster", "prod-cluster-01")
		record.WithDot("infra.kubernetes.namespace", "ecommerce")
		record.WithDot("infra.kubernetes.pod", "api-deployment-abc123-xyz789")

		// Monitoring context
		record.WithDot("monitoring.trace_id", "trace-"+time.Now().Format("20060102150405"))
		record.WithDot("monitoring.span_id", "span-"+time.Now().Format("150405"))
		record.WithDot("monitoring.parent_span_id", nil)

		return record
	})

	complexLogger.Info("Complex nested logging",
		"business.order.id", "order-789",
		"business.customer.tier", "premium",
		"business.payment.method", "credit_card",
	)
}
