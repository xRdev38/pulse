package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
)

// PersistFilter — INSERT TimescaleDB avec Circuit Breaker (Ch.4 Availability)
func PersistFilter(db *pgxpool.Pool, cb *gobreaker.CircuitBreaker) FilterFunc {
	return func(ctx context.Context, p *MetricPayload) error {
		_, err := cb.Execute(func() (interface{}, error) {
			conn, err := db.Acquire(ctx)
			if err != nil {
				return nil, err
			}
			defer conn.Release()

			// RLS : SET LOCAL app.tenant_id isole les données par tenant (Ch.8)
			if _, err = conn.Exec(ctx, "SET LOCAL app.tenant_id = $1", p.TenantID); err != nil {
				return nil, fmt.Errorf("set tenant: %w", err)
			}

			tags, _ := json.Marshal(p.Tags)
			_, err = conn.Exec(ctx,
				`INSERT INTO metrics (time, tenant_id, name, value, tags)
				 VALUES ($1, $2, $3, $4, $5)`,
				time.Unix(0, p.TimestampNs),
				p.TenantID,
				p.NormalizedName,
				p.Value,
				tags,
			)
			return nil, err
		})

		if err != nil {
			if err == gobreaker.ErrOpenState {
				return fmt.Errorf("circuit breaker open: postgres unavailable")
			}
			return fmt.Errorf("persist: %w", err)
		}
		return nil
	}
}

// PublishFilter — publie dans RabbitMQ pour les consommateurs downstream (Ch.13 Broker)
func PublishFilter(mq *amqp.Connection) FilterFunc {
	return func(ctx context.Context, p *MetricPayload) error {
		ch, err := mq.Channel()
		if err != nil {
			return fmt.Errorf("amqp channel: %w", err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare("metrics.ingested", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("declare queue: %w", err)
		}

		body, _ := json.Marshal(map[string]any{
			"tenant_id": p.TenantID,
			"name":      p.NormalizedName,
			"value":     p.Value,
			"tags":      p.Tags,
			"timestamp": p.TimestampNs,
		})

		return ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	}
}
