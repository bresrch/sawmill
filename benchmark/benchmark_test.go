package sawmill

import (
	"io"
	"log"
	"log/slog"
	"testing"
	"time"

	"github.com/bresrch/sawmill"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Shared discard writer for fair comparison
	discardWriter = io.Discard

	// Sample data for structured logging benchmarks
	userID    = 12345
	userName  = "john.doe@example.com"
	requestID = "req-abc123-def456"
	duration  = 250 * time.Millisecond
	timestamp = time.Now()

	// Complex struct for advanced benchmarks
	complexData = struct {
		User     UserInfo               `json:"user"`
		Request  RequestInfo            `json:"request"`
		Response ResponseInfo           `json:"response"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		User: UserInfo{
			ID:       12345,
			Email:    "john.doe@example.com",
			Name:     "John Doe",
			Role:     "admin",
			LastSeen: timestamp,
		},
		Request: RequestInfo{
			ID:        "req-abc123-def456",
			Method:    "POST",
			Path:      "/api/v1/users",
			UserAgent: "Mozilla/5.0 (compatible; benchmark)",
			IP:        "192.168.1.100",
		},
		Response: ResponseInfo{
			StatusCode: 200,
			Size:       1024,
			Duration:   duration,
		},
		Metadata: map[string]interface{}{
			"trace_id":    "trace-xyz789",
			"span_id":     "span-123",
			"environment": "production",
			"version":     "1.2.3",
		},
	}
)

type UserInfo struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	LastSeen time.Time `json:"last_seen"`
}

type RequestInfo struct {
	ID        string `json:"id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}

type ResponseInfo struct {
	StatusCode int           `json:"status_code"`
	Size       int           `json:"size"`
	Duration   time.Duration `json:"duration"`
}

// Setup functions for each logger
func setupSawmill() sawmill.Logger {
	options := sawmill.NewHandlerOptions().
		WithWriter(discardWriter).
		WithLevel(sawmill.LevelInfo)
	handler := sawmill.NewJSONHandler(options)
	return sawmill.New(handler)
}

func setupSawmillDebug() sawmill.Logger {
	options := sawmill.NewHandlerOptions().
		WithWriter(discardWriter).
		WithLevel(sawmill.LevelDebug)
	handler := sawmill.NewJSONHandler(options)
	return sawmill.New(handler)
}

func setupSawmillWarn() sawmill.Logger {
	options := sawmill.NewHandlerOptions().
		WithWriter(discardWriter).
		WithLevel(sawmill.LevelWarn)
	handler := sawmill.NewJSONHandler(options)
	return sawmill.New(handler)
}

func setupStdlib() *log.Logger {
	return log.New(discardWriter, "", log.LstdFlags)
}

func setupSlog() *slog.Logger {
	return slog.New(slog.NewJSONHandler(discardWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func setupLogrus() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(discardWriter)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

func setupZap() *zap.Logger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"/dev/null"}
	config.ErrorOutputPaths = []string{"/dev/null"}
	logger, _ := config.Build()
	return logger
}

func setupZapSugar() *zap.SugaredLogger {
	return setupZap().Sugar()
}

// Simple message benchmarks
func BenchmarkSimpleMessage(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmill()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple log message")
		}
	})

	b.Run("Stdlib", func(b *testing.B) {
		logger := setupStdlib()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Println("Simple log message")
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple log message")
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple log message")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		logger := setupZap()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple log message")
		}
	})

	b.Run("ZapSugar", func(b *testing.B) {
		logger := setupZapSugar()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Simple log message")
		}
	})
}

// Structured logging benchmarks
func BenchmarkStructuredLogging(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmill()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("User request processed",
				"user_id", userID,
				"user_name", userName,
				"request_id", requestID,
				"duration", duration,
				"timestamp", timestamp,
			)
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("User request processed",
				"user_id", userID,
				"user_name", userName,
				"request_id", requestID,
				"duration", duration,
				"timestamp", timestamp,
			)
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"user_name":  userName,
				"request_id": requestID,
				"duration":   duration,
				"timestamp":  timestamp,
			}).Info("User request processed")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		logger := setupZap()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("User request processed",
				zap.Int("user_id", userID),
				zap.String("user_name", userName),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.Time("timestamp", timestamp),
			)
		}
	})

	b.Run("ZapSugar", func(b *testing.B) {
		logger := setupZapSugar()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infow("User request processed",
				"user_id", userID,
				"user_name", userName,
				"request_id", requestID,
				"duration", duration,
				"timestamp", timestamp,
			)
		}
	})
}

// Complex struct logging benchmarks
func BenchmarkComplexStructLogging(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmill()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Complex operation completed",
				"operation", "user_creation",
				"data", complexData,
			)
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Complex operation completed",
				"operation", "user_creation",
				"data", complexData,
			)
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"operation": "user_creation",
				"data":      complexData,
			}).Info("Complex operation completed")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		logger := setupZap()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Complex operation completed",
				zap.String("operation", "user_creation"),
				zap.Any("data", complexData),
			)
		}
	})

	b.Run("ZapSugar", func(b *testing.B) {
		logger := setupZapSugar()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infow("Complex operation completed",
				"operation", "user_creation",
				"data", complexData,
			)
		}
	})
}

// High-frequency logging benchmark (simulates hot path)
func BenchmarkHighFrequency(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmillDebug()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("Debug trace", "iteration", i, "mod", i%100)
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("Debug trace", "iteration", i, "mod", i%100)
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		logger.SetLevel(logrus.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"iteration": i,
				"mod":       i % 100,
			}).Debug("Debug trace")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		config := zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"/dev/null"}
		config.ErrorOutputPaths = []string{"/dev/null"}
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		logger, _ := config.Build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("Debug trace",
				zap.Int("iteration", i),
				zap.Int("mod", i%100),
			)
		}
	})
}

// Level-disabled logging benchmark (important for performance)
func BenchmarkDisabledLevel(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmillWarn() // Debug level disabled
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("This should not be logged", "data", complexData)
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := slog.New(slog.NewJSONHandler(discardWriter, &slog.HandlerOptions{
			Level: slog.LevelWarn, // Debug level disabled
		}))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("This should not be logged", "data", complexData)
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		logger.SetLevel(logrus.WarnLevel) // Debug level disabled
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithField("data", complexData).Debug("This should not be logged")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"/dev/null"}
		config.ErrorOutputPaths = []string{"/dev/null"}
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel) // Debug level disabled
		logger, _ := config.Build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("This should not be logged", zap.Any("data", complexData))
		}
	})
}

// Allocation benchmarks
func BenchmarkAllocations(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmill()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Allocation test",
				"user_id", userID,
				"request_id", requestID,
				"duration", duration,
			)
		}
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Allocation test",
				"user_id", userID,
				"request_id", requestID,
				"duration", duration,
			)
		}
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"request_id": requestID,
				"duration":   duration,
			}).Info("Allocation test")
		}
	})

	b.Run("Zap", func(b *testing.B) {
		logger := setupZap()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("Allocation test",
				zap.Int("user_id", userID),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
			)
		}
	})
}

// Concurrent logging benchmark
func BenchmarkConcurrent(b *testing.B) {
	b.Run("Sawmill", func(b *testing.B) {
		logger := setupSawmill()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("Concurrent logging test",
					"goroutine", "worker",
					"timestamp", time.Now(),
				)
			}
		})
	})

	b.Run("Slog", func(b *testing.B) {
		logger := setupSlog()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("Concurrent logging test",
					"goroutine", "worker",
					"timestamp", time.Now(),
				)
			}
		})
	})

	b.Run("Logrus", func(b *testing.B) {
		logger := setupLogrus()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(logrus.Fields{
					"goroutine": "worker",
					"timestamp": time.Now(),
				}).Info("Concurrent logging test")
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		logger := setupZap()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("Concurrent logging test",
					zap.String("goroutine", "worker"),
					zap.Time("timestamp", time.Now()),
				)
			}
		})
	})
}
