package domain

// SteamID represents steam uint64 id.
type SteamID uint64

// Stats is a map of events, key is event, value is amount of entries of the event.
type Stats map[Event]uint16

// PlayerStats is a map of player event entries.
type PlayerStats struct {
	stats map[SteamID]Stats
}

func NewPlayerStats() *PlayerStats {
	return &PlayerStats{
		stats: make(map[SteamID]Stats),
	}
}

// Get returns stats of specific player with steamID.
func (s *PlayerStats) Get(steamID uint64) (Stats, bool) {
	v, ok := s.stats[SteamID(steamID)]
	return v, ok
}

// Incr increments amount of player event entries.
func (s *PlayerStats) Incr(steamID uint64, e Event) {
	_, ok := s.stats[SteamID(steamID)]
	if !ok {
		s.stats[SteamID(steamID)] = make(Stats)
	}

	s.stats[SteamID(steamID)][e] += 1
}
