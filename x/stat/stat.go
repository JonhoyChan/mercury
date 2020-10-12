package stat

import "mercury/x/stat/prometheus"

// Stat interface.
type Stat interface {
	Timing(name string, time int64, extra ...string)
	Incr(name string, extra ...string) // name,ext...,code
	State(name string, val int64, extra ...string)
}

// default stat struct.
var (
	// http
	HTTPServer Stat = prometheus.HTTPServer
	// rpc
	RPCServer Stat = prometheus.RPCServer
)
