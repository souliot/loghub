package metrics

import (
	"net/http"

	"public/libs_go/gateway/metrics/service"
	"public/libs_go/gateway/metrics/system"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
