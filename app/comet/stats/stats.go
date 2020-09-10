package stats

import (
	"expvar"
	"time"

	"github.com/micro/go-micro/v2/web"
)

// Runtime statistics communication channel.
var StatsUpdate chan *varUpdate

type varUpdate struct {
	// Name of the variable to update
	name string
	// Integer value to publish
	count int64
	// Treat the count as an increment as opposite to the final value.
	inc bool
}

// Initialize stats reporting through expvar.
func Init(microWeb web.Service, path string) {
	if path == "" || path == "-" {
		return
	}

	microWeb.Handle(path, expvar.Handler())
	StatsUpdate = make(chan *varUpdate, 1024)

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
	if StatsUpdate != nil {
		select {
		case StatsUpdate <- &varUpdate{name, int64(val), inc}:
		default:
		}
	}
}

// Stop publishing stats.
func Shutdown() {
	if StatsUpdate != nil {
		StatsUpdate <- nil
	}
}

// The go routine which actually publishes stats updates.
func updater() {
	for upd := range StatsUpdate {
		if upd == nil {
			StatsUpdate = nil
			// Dont' care to close the channel.
			break
		}

		// Handle var update
		if ev := expvar.Get(upd.name); ev != nil {
			// Intentional panic if the ev is not *expvar.Int.
			intvar := ev.(*expvar.Int)
			if upd.inc {
				intvar.Add(upd.count)
			} else {
				intvar.Set(upd.count)
			}
		} else {
			panic("stats: update to unknown variable " + upd.name)
		}
	}
}
