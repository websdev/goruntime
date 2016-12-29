package mocks

import "github.com/lyft/goruntime/stats"

type Scope struct{}

func (m *Scope) Scope(name string) stats.Scope        { return &Scope{} }
func (m *Scope) NewCounter(name string) stats.Counter { return &Counter{} }
func (m *Scope) NewGauge(name string) stats.Gauge     { return &Gauge{} }

type Counter struct{}

func (m *Counter) Inc() {}

type Gauge struct{}

func (m *Gauge) Set(uint64) {}
