package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sapliy/fintech-ecosystem/internal/ledger/domain"
	"github.com/sapliy/fintech-ecosystem/internal/ledger/infrastructure"
	"github.com/sapliy/fintech-ecosystem/pkg/database"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
	"github.com/sapliy/fintech-ecosystem/pkg/observability"
	pb "github.com/sapliy/fintech-ecosystem/proto/ledger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapliy/fintech-ecosystem/pkg/messaging"
	"github.com/sapliy/fintech-ecosystem/pkg/monitoring"
	"google.golang.org/grpc"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@127.0.0.1:5435/ledger?sslmode=disable"
	}

	db, err := database.Connect(dsn)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
	} else {
		log.Println("Database connection established")

		// Run automated migrations
		if err := database.Migrate(db, "ledger", "migrations/ledger"); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	}
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Failed to close DB: %v", err)
			}
		}()
	}

	// Initialize Redis for Caching
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis connection failed in Ledger: %v", err)
	}

	// Layered Architecture Setup
	sqlRepo := infrastructure.NewSQLRepository(db)
	repo := infrastructure.NewCachedRepository(sqlRepo, rdb)
	metrics := &infrastructure.PrometheusMetrics{}
	service := domain.NewLedgerService(repo, metrics)

	// Initialize Tracer
	shutdown, err := observability.InitTracer(context.Background(), observability.Config{
		ServiceName:    "ledger",
		ServiceVersion: "0.1.0",
		Endpoint:       os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Environment:    "production",
	})
	if err != nil {
		log.Printf("Failed to init tracer: %v", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				log.Printf("Failed to shutdown tracer: %v", err)
			}
		}()
	}

	// Start Kafka Consumer for Event Sourcing
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	brokers := strings.Split(kafkaBrokers, ",")
	go StartKafkaConsumer(brokers, service)

	// Start Outbox Publisher for Reliable Event Delivery
	ledgerProducer := messaging.NewKafkaProducer(brokers, "ledger-events")
	publisher := infrastructure.NewOutboxPublisher(repo, ledgerProducer, 2*time.Second)
	go publisher.Start(context.Background())

	handler := &LedgerHandler{service: service}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "active",
			"service": "ledger",
		})
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/accounts", handler.CreateAccount)

	// Simple routing for /accounts/{id}
	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetAccount(w, r)
			return
		}
		jsonutil.WriteErrorJSON(w, "Not Found")
	})

	mux.HandleFunc("/transactions", handler.RecordTransaction)
	mux.HandleFunc("/bulk-transactions", handler.BulkRecordTransactions)

	log.Println("Ledger service HTTP starting on :8083")

	// Wrap handler with OpenTelemetry and Prometheus
	otelHandler := otelhttp.NewHandler(mux, "ledger-request")
	promHandler := monitoring.PrometheusMiddleware(otelHandler)

	go func() {
		if err := http.ListenAndServe(":8083", promHandler); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen for gRPC: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(monitoring.UnaryServerInterceptor("ledger")),
	)
	pb.RegisterLedgerServiceServer(s, NewLedgerGRPCServer(service))

	log.Println("Ledger service gRPC starting on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
