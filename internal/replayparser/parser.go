package replayparser

import (
	"io"

	"github.com/ssssargsian/uniplay/internal/domain"

	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
)

// parser is a wrapper around demoinfocs.Parser.
// ONE parser must be used for ONE replay.
type parser struct {
	demoinfocs.Parser

	metrics       *playerMetrics
	weaponMetrics *weaponMetrics
	match         domain.Match

	isKnifeRound bool
}

type parseResult struct {
	Metrics       *playerMetrics
	WeaponMetrics *weaponMetrics
	Match         domain.Match
}

func New(r io.Reader) *parser {
	return &parser{
		demoinfocs.NewParser(r),
		newPlayerMetrics(),
		newWeaponMetrics(),
		domain.Match{},
		false,
	}
}

// collectStats detects if stats can be collected to prevent collection of stats on knife or warmup rounds.
// return false if current round is knife round or match is not started.
func (p *parser) collectStats(gs demoinfocs.GameState) bool {
	if p.isKnifeRound || !gs.IsMatchStarted() {
		return false
	}

	return true
}

// detectKnifeRound sets isKnifeRound to true if any player has no secondary weapon and first slot is a knife.
func (p *parser) detectKnifeRound() {
	p.isKnifeRound = false

	for _, player := range p.GameState().TeamCounterTerrorists().Members() {
		weapons := player.Weapons()
		if len(player.Weapons()) == 1 && weapons[0].Type == common.EqKnife {
			p.isKnifeRound = true
			break
		}
	}
}

// handleKills counts all kills and weapon kills.
func (p *parser) handleKills() {
	p.RegisterEventHandler(func(e events.Kill) {
		if !p.collectStats(p.GameState()) {
			return
		}

		var weapon string
		if e.Weapon != nil {
			weapon = e.Weapon.String()
		}

		if e.Victim != nil {
			// death amount FROM weapon
			p.metrics.incr(e.Victim.SteamID64, domain.MetricDeath)
			p.weaponMetrics.incr(e.Victim.SteamID64, weapon, domain.MetricDeath)
		}

		if e.Killer != nil {
			// kill amount
			p.metrics.incr(e.Killer.SteamID64, domain.MetricKill)
			p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricKill)

			switch {
			// headshot kill amount
			case e.IsHeadshot:
				p.metrics.incr(e.Killer.SteamID64, domain.MetricHSKill)
				p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricHSKill)

			// blind kill amount
			case e.AttackerBlind:
				p.metrics.incr(e.Killer.SteamID64, domain.MetricBlindKill)
				p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricBlindKill)

			// wallbang kill amount
			case e.IsWallBang():
				p.metrics.incr(e.Killer.SteamID64, domain.MetricWallbangKill)
				p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricWallbangKill)

			// noscope kill amount
			case e.NoScope:
				p.metrics.incr(e.Killer.SteamID64, domain.MetricNoScopeKill)
				p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricNoScopeKill)

			// through smoke kill amount
			case e.ThroughSmoke:
				p.metrics.incr(e.Killer.SteamID64, domain.MetricThroughSmokeKill)
				p.weaponMetrics.incr(e.Killer.SteamID64, weapon, domain.MetricThroughSmokeKill)
			}
		}

		if e.Assister != nil {
			// assist total amount
			p.metrics.incr(e.Assister.SteamID64, domain.MetricAssist)

			// flashbang assist amount
			if e.AssistedFlash {
				p.metrics.incr(e.Assister.SteamID64, domain.MetricFlashbangAssist)
			}
		}
	})
}

// handleScoreUpdate updates match teams score on ScoreUpdated event.
func (p *parser) handleScoreUpdate() {
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		switch e.TeamState.Team() {
		case common.TeamCounterTerrorists:
			p.match.Team1.SetAll(e.TeamState.ClanName(), e.TeamState.Flag(), e.TeamState.Score())
		case common.TeamTerrorists:
			p.match.Team2.SetAll(e.TeamState.ClanName(), e.TeamState.Flag(), e.TeamState.Score())
		}
	})
}

// handlePlayerHurt calculates metrics for taken, dealt damage and for taken and dealt damage by a weapon.
func (p *parser) handlePlayerHurt() {
	p.RegisterEventHandler(func(e events.PlayerHurt) {
		if !p.collectStats(p.GameState()) {
			return
		}

		var weapon string
		if e.Weapon != nil {
			weapon = e.Weapon.String()
		}

		if e.Attacker != nil {
			// dealt damage
			p.metrics.add(e.Attacker.SteamID64, domain.MetricDamageDealt, e.HealthDamage)
			p.weaponMetrics.add(e.Attacker.SteamID64, weapon, domain.MetricDamageDealt, e.HealthDamage)
		}

		if e.Player != nil {
			// taken damage
			p.metrics.add(e.Player.SteamID64, domain.MetricDamageTaken, e.HealthDamage)
			p.weaponMetrics.add(e.Player.SteamID64, weapon, domain.MetricDamageTaken, e.HealthDamage)
		}
	})
}

// handleBomdEvent counts number of planted and defused bombs.
func (p *parser) handleBombEvents() {
	p.RegisterEventHandler(func(e events.BombDefused) {
		if !p.collectStats(p.GameState()) {
			return
		}

		if e.BombEvent.Player != nil {
			p.metrics.incr(e.BombEvent.Player.SteamID64, domain.MetricBombDefused)
		}
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		if !p.collectStats(p.GameState()) {
			return
		}

		if e.BombEvent.Player != nil {
			p.metrics.incr(e.BombEvent.Player.SteamID64, domain.MetricBombPlanted)
		}
	})
}

// handleMVPAnnouncement increments player mvp amount metric on RoundMVPAnnouncement event.
func (p *parser) handleMVPAnnouncement() {
	p.RegisterEventHandler(func(e events.RoundMVPAnnouncement) {
		if !p.collectStats(p.GameState()) {
			return
		}

		if e.Player != nil {
			p.metrics.incr(e.Player.SteamID64, domain.MetricRoundMVPCount)
		}
	})
}

func (p *parser) Parse() (parseResult, error) {
	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		p.detectKnifeRound()
	})

	p.handleKills()
	p.handlePlayerHurt()
	p.handleScoreUpdate()
	p.handleMVPAnnouncement()
	p.handleBombEvents()

	p.RegisterEventHandler(func(e events.AnnouncementWinPanelMatch) {
		p.match.MapName = p.Header().MapName
		p.match.Duration = p.Header().PlaybackTime
	})

	if err := p.ParseToEnd(); err != nil {
		return parseResult{}, err
	}

	return parseResult{
		Metrics:       p.metrics,
		WeaponMetrics: p.weaponMetrics,
		Match:         p.match,
	}, nil
}
