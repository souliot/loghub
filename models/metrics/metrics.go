package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/souliot/gateway/metrics/service"
	"github.com/souliot/gateway/metrics/system"
)

var (
	DefaultMetrics = new(Metrics)
	Handler        http.Handler
)

type Metrics struct{}

func (m *Metrics) Init() {
	r := prometheus.NewRegistry()
	system.RegisterSystemCollector(r)
	service.RegisterServiceCollector(r, &service.RegisterOptions{"loghub"})

	Handler = promhttp.HandlerFor(
		prometheus.Gatherers{r},
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)
	return
}
