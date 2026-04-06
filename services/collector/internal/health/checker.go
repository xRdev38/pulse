package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type check struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Checker — Health endpoint pour Docker + orchestrateurs (Ch.4 Availability)
type Checker struct {
	db     *pgxpool.Pool
	mq     *amqp.Connection
	logger *zap.Logger
}

func NewChecker(db *pgxpool.Pool, mq *amqp.Connection, logger *zap.Logger) *Checker {
	return &Checker{db: db, mq: mq, logger: logger}
}

func (c *Checker) Handle(ctx *gin.Context) {
	checks := map[string]check{
		"postgres": c.postgres(ctx.Request.Context()),
		"rabbitmq": c.rabbitmq(),
	}

	overall := "healthy"
	for _, s := range checks {
		if s.Status == "error" {
			overall = "degraded"
		}
	}

	code := http.StatusOK
	if overall == "degraded" {
		code = http.StatusServiceUnavailable
	}
	ctx.JSON(code, gin.H{"status": overall, "checks": checks})
}

func (c *Checker) postgres(ctx context.Context) check {
	t := time.Now()
	pCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.db.Ping(pCtx); err != nil {
		return check{Status: "error", Error: err.Error()}
	}
	return check{Status: "ok", Latency: time.Since(t).String()}
}

func (c *Checker) rabbitmq() check {
	if c.mq.IsClosed() {
		return check{Status: "error", Error: "connection closed"}
	}
	return check{Status: "ok"}
}
