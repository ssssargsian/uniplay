package domain

import "time"

type Match struct {
	Map      string
	Duration time.Duration
	Team1    MatchTeam
	Team2    MatchTeam
}

type MatchTeam struct {
	Name     string
	FlagCode string
	Score    int
}

func (t *MatchTeam) SetAll(name, flag string, score int) {
	t.Name = name
	t.FlagCode = flag
	t.Score = score
}

func (t *MatchTeam) IsWinner(opponent MatchTeam) bool {
	return t.Score > opponent.Score
}
