package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPServer for http server
	HTTPServer = New().
			WithTimer("http_server_response_time_seconds", []string{"method", "path"}).
			WithState("http_server_state", []string{"method", "name"}).
			WithCounter("http_server_request_total", []string{"method", "code"})
	// RPCServer for rpc server
	RPCServer = New().
			WithTimer("rpc_server_response_time_seconds", []string{"method"}).
			WithCounter("rpc_server_request_total", []string{"method", "code"})
)

// Prometheus struct info
type Prometheus struct {
	timer   *prometheus.HistogramVec
	counter *prometheus.CounterVec
	state   *prometheus.GaugeVec
}

// New creates a Prom instance.
func New() *Prometheus {
	return &Prometheus{}
}

// WithTimer with summary timer
func (p *Prometheus) WithTimer(name string, labels []string) *Prometheus {
	if p == nil || p.timer != nil {
		return p
	}
	p.timer = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name,
			Help: "Description",
		}, labels)
	prometheus.MustRegister(p.timer)
	return p
}

// WithCounter sets counter.
func (p *Prometheus) WithCounter(name string, labels []string) *Prometheus {
	if p == nil || p.counter != nil {
		return p
	}
	p.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: "Description",
		}, labels)
	prometheus.MustRegister(p.counter)
	return p
}

// WithState sets state.
func (p *Prometheus) WithState(name string, labels []string) *Prometheus {
	if p == nil || p.state != nil {
		return p
	}
	p.state = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: "Description",
		}, labels)
	prometheus.MustRegister(p.state)
	return p
}

// Timing log timing information (in milliseconds) without sampling
func (p *Prometheus) Timing(name string, time int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.timer != nil {
		p.timer.WithLabelValues(label...).Observe(float64(time))
	}
}

// Incr increments one stat counter without sampling
func (p *Prometheus) Incr(name string, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Inc()
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Inc()
	}
}

// Decr decrements one stat counter without sampling
func (p *Prometheus) Decr(name string, extra ...string) {
	if p.state != nil {
		label := append([]string{name}, extra...)
		p.state.WithLabelValues(label...).Dec()
	}
}

// State set state
func (p *Prometheus) State(name string, v int64, extra ...string) {
	if p.state != nil {
		label := append([]string{name}, extra...)
		p.state.WithLabelValues(label...).Set(float64(v))
	}
}

// Add add count    v must > 0
func (p *Prometheus) Add(name string, v int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Add(float64(v))
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Add(float64(v))
	}
}
