package pipeline

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// MetricPayload est la donnée qui traverse le pipeline
type MetricPayload struct {
	Name           string
	Value          float64
	Tags           map[string]string
	TenantID       string
	NormalizedName string
	TimestampNs    int64
	Metadata       map[string]any
}

// FilterFunc est la signature d'un filtre
type FilterFunc func(ctx context.Context, p *MetricPayload) error

// Pipeline — Pipe-and-Filter pattern (Bass/Clements/Kazman Ch.13)
type Pipeline struct {
	filters []FilterFunc
	logger  *zap.Logger
}

func New(logger *zap.Logger) *Pipeline {
	return &Pipeline{logger: logger}
}

// Add ajoute un filtre en bout de chaîne
func (p *Pipeline) Add(f FilterFunc) *Pipeline {
	p.filters = append(p.filters, f)
	return p
}

// Execute fait passer le payload par tous les filtres (fail-fast)
func (p *Pipeline) Execute(ctx context.Context, payload *MetricPayload) error {
	for i, filter := range p.filters {
		if err := filter(ctx, payload); err != nil {
			p.logger.Warn("filter failed",
				zap.Int("step", i+1),
				zap.String("metric", payload.Name),
				zap.Error(err),
			)
			return fmt.Errorf("filter[%d]: %w", i+1, err)
		}
	}
	return nil
}
