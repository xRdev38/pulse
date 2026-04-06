package circuit

import (
	"os"
	"strconv"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// NewPostgresBreaker — Circuit Breaker sur PostgreSQL (Ch.4 Availability tactic)
//
// États : Closed (normal) → Open (panne détectée) → Half-Open (test de reprise)
// Équivalent Node.js : opossum (npm)
func NewPostgresBreaker(logger *zap.Logger) *gobreaker.CircuitBreaker {
	maxReq, _  := strconv.Atoi(env("CB_MAX_REQUESTS", "3"))
	interval, _ := strconv.Atoi(env("CB_INTERVAL_SECS", "10"))
	timeout, _  := strconv.Atoi(env("CB_TIMEOUT_SECS", "30"))

	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "postgres",
		MaxRequests: uint32(maxReq),
		Interval:    time.Duration(interval) * time.Second,
		Timeout:     time.Duration(timeout) * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Warn("circuit breaker state changed",
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
