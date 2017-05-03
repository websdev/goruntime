package snapshot

import "github.com/lyft/goruntime/snapshot/entry"

// Implementation of Snapshot for the nilLoaderImpl.
type Mock struct {
	*Snapshot

	entries map[string]*entry.Entry
}

func NewMock() (s *Mock) {
	s = &Mock{
		Snapshot: New(),
		entries:  make(map[string]*entry.Entry),
	}

	return
}

func (m *Mock) SetEnabled(key string) *Mock {
	m.Snapshot.entries[key] = entry.New(key, 0, true)

	return m
}

func (m *Mock) SetDisabled(key string) *Mock {
	m.Snapshot.entries[key] = entry.New(key, 0, false)

	return m
}

func (m *Mock) SetEntry(key string, val string) *Mock {
	m.Snapshot.entries[key] = entry.New(val, 0, false)

	return m
}

func (m *Mock) FeatureEnabled(key string, defaultValue uint64) bool {
	if e, ok := m.Snapshot.entries[key]; ok {
		return e.Uint64Valid
	}

	return false
}
