package pipeline

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{0,127}$`)

// ValidateFilter — vérifie le contrat d'entrée (Ch.4 : fail fast)
func ValidateFilter() FilterFunc {
	return func(ctx context.Context, p *MetricPayload) error {
		if p.Name == "" {
			return fmt.Errorf("name is required")
		}
		if !validName.MatchString(p.Name) {
			return fmt.Errorf("invalid name %q: must match [a-zA-Z][a-zA-Z0-9._-]", p.Name)
		}
		if p.TenantID == "" {
			return fmt.Errorf("tenant_id is required")
		}
		if p.Value != p.Value { // NaN
			return fmt.Errorf("value cannot be NaN")
		}
		if len(p.Tags) > 32 {
			return fmt.Errorf("too many tags (max 32, got %d)", len(p.Tags))
		}
		return nil
	}
}

// NormalizeFilter — standardise noms et tags
func NormalizeFilter() FilterFunc {
	return func(ctx context.Context, p *MetricPayload) error {
		p.NormalizedName = strings.ToLower(strings.TrimSpace(p.Name))

		norm := make(map[string]string, len(p.Tags))
		for k, v := range p.Tags {
			key := strings.ToLower(strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
					return r
				}
				return '_'
			}, strings.TrimSpace(k)))
			norm[key] = strings.TrimSpace(v)
		}
		p.Tags = norm

		if p.Metadata == nil {
			p.Metadata = make(map[string]any)
		}
		return nil
	}
}

// EnrichFilter — ajoute le timestamp serveur (évite la dérive d'horloge client)
func EnrichFilter() FilterFunc {
	return func(ctx context.Context, p *MetricPayload) error {
		p.TimestampNs = time.Now().UnixNano()
		p.Metadata["ingested_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		if p.Tags == nil {
			p.Tags = make(map[string]string)
		}
		p.Tags["_tenant"] = p.TenantID
		return nil
	}
}
