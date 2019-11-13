package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"math"
)

func (s *Store) GetDirtyEpochStats() *EpochStats {
	return s.GetEpochStats(idx.Epoch(math.MaxInt32))
}

func (s *Store) GetEpochStats(epoch idx.Epoch) *EpochStats {
	key := epoch.Bytes()

	// Get data from LRU cache first.
	if s.cache.EpochStats != nil {
		if c, ok := s.cache.EpochStats.Get(string(key)); ok {
			if b, ok := c.(EpochStats); ok {
				return &b
			}
		}
	}

	w, _ := s.get(s.table.EpochStats, key, &EpochStats{}).(*EpochStats)

	// Add to LRU cache.
	if w != nil && s.cache.EpochStats != nil {
		s.cache.EpochStats.Add(string(key), *w)
	}

	return w
}

func (s *Store) SetDirtyEpochStats(value EpochStats) {
	s.SetEpochStats(idx.Epoch(math.MaxInt32), value)
}

func (s *Store) SetEpochStats(epoch idx.Epoch, value EpochStats) {
	key := epoch.Bytes()

	s.set(s.table.EpochStats, key, value)

	// Add to LRU cache.
	if s.cache.EpochStats != nil {
		s.cache.EpochStats.Add(string(key), value)
	}
}
