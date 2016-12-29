// stats defines the minimal behavior needed in order to export application metrics
// from runtime. Adapters need only implement the `Scope` interface for all internal
// counter and gauge metrics to be emitted. The `mocks` package provides a nullified
// `Scope` implementation if no metrics are needed within an application.
package stats

// Counter is a definition of a way in which to increment metrics by name
type Counter interface {
	// Inc increments a counter value
	Inc()
}

// Gauge is a definition of a way in which to set metrics by name
type Gauge interface {
	// Set applies its argument to a gauge
	Set(uint64)
}

// Scope defines an implementation of top level stats within `runtime`. `Counters` and
// `Gauges` are initialized according to their implementations.
type Scope interface {
	// Scope applies a new namespace nesting to the parent `Scope`.
	Scope(name string) Scope

	// NewCounter initializes a new `Counter`
	NewCounter(name string) Counter

	// NewGauge initializes a new `Gauge`
	NewGauge(name string) Gauge
}
