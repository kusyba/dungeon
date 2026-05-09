package player

import (
	"time"
)

type Player struct {
	ID                 int
	Registered         bool
	InDungeon          bool
	CurrentFloor       int
	MonstersKilled     int
	FloorsCompleted    []bool
	FloorStartTime     time.Time
	FloorClearDuration []time.Duration
	BossDefeated       bool
	BossStartTime      time.Time
	BossKillDuration   time.Duration
	EnterTime          time.Time
	ExitTime           time.Time
	LastEventTime      time.Time
	Health             int
	Disqualified       bool
	Dead               bool
	CannotContinue     bool
}

func NewPlayer(id int, totalFloors int) *Player {
	return &Player{
		ID:                 id,
		Health:             100,
		FloorsCompleted:    make([]bool, totalFloors+2),
		FloorClearDuration: make([]time.Duration, totalFloors+2),
	}
}

func (p *Player) Register() bool {
	if !p.Registered {
		p.Registered = true
		return true
	}
	return false
}

func (p *Player) EnterDungeon(enterTime time.Time) {
	p.InDungeon = true
	p.EnterTime = enterTime
	p.CurrentFloor = 1
	p.MonstersKilled = 0
	p.FloorStartTime = enterTime
}

func (p *Player) KillMonster(eventTime time.Time, monstersPerFloor int) bool {
	p.MonstersKilled++
	if p.MonstersKilled >= monstersPerFloor {
		p.FloorClearDuration[p.CurrentFloor] = eventTime.Sub(p.FloorStartTime)
		p.FloorsCompleted[p.CurrentFloor] = true
		p.MonstersKilled = 0
		return true
	}
	return false
}

func (p *Player) CanMoveToNextFloor(totalFloors int) bool {
	if p.CurrentFloor > totalFloors {
		return false
	}
	return p.FloorsCompleted[p.CurrentFloor]
}

func (p *Player) MoveToNextFloor(eventTime time.Time) {
	p.CurrentFloor++
	p.FloorStartTime = eventTime
}

func (p *Player) EnterBossFloor(eventTime time.Time, bossFloor int) {
	p.CurrentFloor = bossFloor
	p.BossStartTime = eventTime
}

func (p *Player) KillBoss(eventTime time.Time) {
	p.BossDefeated = true
	p.BossKillDuration = eventTime.Sub(p.BossStartTime)
}

func (p *Player) LeaveDungeon(exitTime time.Time) {
	p.InDungeon = false
	p.ExitTime = exitTime
}

func (p *Player) SetCannotContinue(exitTime time.Time) {
	p.CannotContinue = true
	p.InDungeon = false
	p.ExitTime = exitTime
}

func (p *Player) RestoreHealth(amount int) {
	p.Health += amount
	if p.Health > 100 {
		p.Health = 100
	}
}

func (p *Player) TakeDamage(amount int) bool {
	p.Health -= amount
	if p.Health <= 0 {
		p.Health = 0
		p.Dead = true
		p.InDungeon = false
		return true
	}
	return false
}

func (p *Player) IsActive() bool {
	return !p.Disqualified && !p.Dead && !p.CannotContinue
}

func (p *Player) IsSuccess(totalFloors int) bool {
	// Проверяем, что босс побеждён
	if !p.BossDefeated {
		return false
	}
	// Проверяем, что все обычные этажи очищены
	for i := 1; i <= totalFloors; i++ {
		if !p.FloorsCompleted[i] {
			return false
		}
	}
	return true
}

func (p *Player) GetTotalTime() time.Duration {
	endTime := p.ExitTime
	if endTime.IsZero() && p.InDungeon {
		endTime = p.LastEventTime
	}
	if endTime.IsZero() {
		endTime = p.LastEventTime
	}
	if !endTime.IsZero() && !p.EnterTime.IsZero() {
		return endTime.Sub(p.EnterTime)
	}
	return 0
}

func (p *Player) GetAverageClearTime(totalFloors int) time.Duration {
	clearedCount := 0
	var sumNanos int64 = 0
	for i := 1; i <= totalFloors; i++ {
		if p.FloorsCompleted[i] {
			clearedCount++
			sumNanos += int64(p.FloorClearDuration[i])
		}
	}
	if clearedCount > 0 {
		return time.Duration(sumNanos / int64(clearedCount))
	}
	return 0
}
