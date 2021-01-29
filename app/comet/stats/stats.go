package stats

import (
	"expvar"
	"net/http"
	"time"
)

// Runtime statistics communication channel.
var update chan *varUpdate

type varUpdate struct {
	// Name of the variable to update
	name string
	// Integer value to publish
	count int64
	// Treat the count as an increment as opposite to the final value.
	inc bool
}

var Handler http.Handler

// Initialize stats reporting through expvar.
func init() {
	Handler = expvar.Handler()
	update = make(chan *varUpdate, 1024)

	start := time.Now()
	expvar.Publish("Uptime", expvar.Func(func() interface{} {
		return time.Since(start).Seconds()
	}))

	go updater()
}

// Register integer variable. Don't check for initialization.
func RegisterInt(name string) {
	expvar.Publish(name, new(expvar.Int))
}

// Async publish int variable.
func Set(name string, val int, inc bool) {
	if update != nil {
		select {
		case update <- &varUpdate{name, int64(val), inc}:
		default:
		}
	}
}

// Stop publishing stats.
func Shutdown() {
	if update != nil {
		update <- nil
	}
}

// The go routine which actually publishes stats updates.
func updater() {
	for upd := range update {
		if upd == nil {
			update = nil
			// Dont' care to close the channel.
			break
		}

		// Handle var update
		if ev := expvar.Get(upd.name); ev != nil {
			// Intentional panic if the ev is not *expvar.Int.
			intVar := ev.(*expvar.Int)
			if upd.inc {
				intVar.Add(upd.count)
			} else {
				intVar.Set(upd.count)
			}
		} else {
			panic("stats: update to unknown variable " + upd.name)
		}
	}
}
