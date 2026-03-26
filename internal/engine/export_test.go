package engine

import (
	"context"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/rule"
	"github.com/carlosonunez/status/internal/setter"
)

// NewForTest creates an Engine suitable for unit-testing processGroup behavior.
// The ticker factory is left as the real default; use SetTickerFactory to
// override it for Run tests.
func NewForTest(sr *setter.Registry, now func() time.Time) *Engine {
	return &Engine{
		setters:   sr,
		now:       now,
		newTicker: defaultTicker,
	}
}

// ProcessGroup exposes processGroup for direct testing without a running ticker.
func (e *Engine) ProcessGroup(
	ctx context.Context,
	g getter.EventGetter,
	p params.Params,
	rules []*rule.Rule,
) {
	e.processGroup(ctx, getterGroup{getter: g, params: p, rules: rules})
}

// SetTickerFactory overrides the ticker factory used by Run. Tests use this to
// inject a never-firing or controlled channel.
func SetTickerFactory(e *Engine, f func(time.Duration) (<-chan time.Time, func())) {
	e.newTicker = f
}
