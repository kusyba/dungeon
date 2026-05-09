package processor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"dungeon/internal/config"
	"dungeon/internal/player"
)

type Processor struct {
	cfg       *config.Config
	players   map[int]*player.Player
	openTime  time.Time
	closeTime time.Time
	output    []string
}

func NewProcessor(cfg *config.Config, openTime, closeTime time.Time) *Processor {
	return &Processor{
		cfg:       cfg,
		players:   make(map[int]*player.Player),
		openTime:  openTime,
		closeTime: closeTime,
		output:    make([]string, 0),
	}
}

func (p *Processor) ProcessEvent(eventLine string) error {
	parts := strings.Fields(eventLine)
	if len(parts) < 3 {
		return fmt.Errorf("invalid event format: %s", eventLine)
	}

	timeWithBrackets := parts[0]
	timeWithoutBrackets := strings.Trim(parts[0], "[]")
	
	eventID, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid event ID: %s", parts[1])
	}
	
	playerID, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid player ID: %s", parts[2])
	}

	eventTime, err := time.Parse("15:04:05", timeWithoutBrackets)
	if err != nil {
		return fmt.Errorf("invalid time format: %s", parts[0])
	}

	extraParam := ""
	if len(parts) > 3 {
		extraParam = parts[3]
	}

	pl := p.getOrCreatePlayer(playerID)

	if eventTime.After(p.closeTime) && pl.InDungeon && !pl.Dead && !pl.CannotContinue {
		pl.LeaveDungeon(p.closeTime)
	}

	if eventTime.Before(p.openTime) && eventID != 1 {
		if !pl.Disqualified {
			pl.Disqualified = true
			p.output = append(p.output, fmt.Sprintf("%s Player [%d] is disqualified", timeWithBrackets, playerID))
		}
		return nil
	}

	if (pl.Disqualified || pl.Dead || pl.CannotContinue) && eventID != 1 {
		return nil
	}

	var outputLine string

	switch eventID {
	case 1:
		outputLine = p.handleRegister(timeWithBrackets, playerID, pl)
	case 2:
		outputLine = p.handleEnter(timeWithBrackets, playerID, pl, eventTime)
	case 3:
		outputLine = p.handleKillMonster(timeWithBrackets, playerID, pl, eventTime)
	case 4:
		outputLine = p.handleNextFloor(timeWithBrackets, playerID, pl, eventTime)
	case 5:
		outputLine = p.handlePrevFloor(timeWithBrackets, playerID, pl)
	case 6:
		outputLine = p.handleEnterBoss(timeWithBrackets, playerID, pl, eventTime)
	case 7:
		outputLine = p.handleKillBoss(timeWithBrackets, playerID, pl, eventTime)
	case 8:
		outputLine = p.handleLeave(timeWithBrackets, playerID, pl, eventTime)
	case 9:
		outputLine = p.handleCannotContinue(timeWithBrackets, playerID, pl, eventTime, extraParam)
	case 10:
		outputLine, err = p.handleRestoreHealth(timeWithBrackets, playerID, pl, extraParam)
		if err != nil {
			return err
		}
	case 11:
		outputLine, err = p.handleDamage(timeWithBrackets, playerID, pl, extraParam)
		if err != nil {
			return err
		}
	}

	if outputLine != "" {
		p.output = append(p.output, outputLine)
	}

	pl.LastEventTime = eventTime
	return nil
}

func (p *Processor) getOrCreatePlayer(id int) *player.Player {
	if _, exists := p.players[id]; !exists {
		p.players[id] = player.NewPlayer(id, p.cfg.Floors)
	}
	return p.players[id]
}

func (p *Processor) handleRegister(timeStr string, playerID int, pl *player.Player) string {
	pl.Registered = true
	pl.Disqualified = false
	pl.Dead = false
	pl.CannotContinue = false
	pl.InDungeon = false
	pl.Health = 100
	pl.BossDefeated = false
	pl.MonstersKilled = 0
	pl.CurrentFloor = 0
	pl.EnterTime = time.Time{}
	pl.ExitTime = time.Time{}
	pl.BossKillDuration = 0
	
	for i := range pl.FloorsCompleted {
		pl.FloorsCompleted[i] = false
	}
	for i := range pl.FloorClearDuration {
		pl.FloorClearDuration[i] = 0
	}
	
	return fmt.Sprintf("%s Player [%d] registered", timeStr, playerID)
}

func (p *Processor) handleEnter(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.Registered {
		return fmt.Sprintf("%s Player [%d] is disqualified", timeStr, playerID)
	}
	if eventTime.Before(p.openTime) || eventTime.After(p.closeTime) {
		pl.Disqualified = true
		return fmt.Sprintf("%s Player [%d] is disqualified", timeStr, playerID)
	}
	if !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [2]", timeStr, playerID)
	}
	if !pl.InDungeon {
		pl.EnterDungeon(eventTime)
		pl.MonstersKilled = 0
		pl.FloorStartTime = eventTime
		return fmt.Sprintf("%s Player [%d] entered the dungeon", timeStr, playerID)
	}
	return ""
}

func (p *Processor) handleKillMonster(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [3]", timeStr, playerID)
	}
	if pl.CurrentFloor > p.cfg.Floors {
		return fmt.Sprintf("%s Player [%d] makes imposible move [3]", timeStr, playerID)
	}
	if pl.FloorsCompleted[pl.CurrentFloor] {
		return fmt.Sprintf("%s Player [%d] makes imposible move [3]", timeStr, playerID)
	}
	if pl.MonstersKilled >= p.cfg.Monsters {
		return fmt.Sprintf("%s Player [%d] makes imposible move [3]", timeStr, playerID)
	}

	pl.MonstersKilled++
	
	if pl.MonstersKilled >= p.cfg.Monsters {
		pl.FloorClearDuration[pl.CurrentFloor] = eventTime.Sub(pl.FloorStartTime)
		pl.FloorsCompleted[pl.CurrentFloor] = true
		pl.MonstersKilled = 0
	}
	
	return fmt.Sprintf("%s Player [%d] killed the monster", timeStr, playerID)
}

func (p *Processor) handleNextFloor(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [4]", timeStr, playerID)
	}
	if !pl.CanMoveToNextFloor(p.cfg.Floors) {
		return fmt.Sprintf("%s Player [%d] makes imposible move [4]", timeStr, playerID)
	}
	pl.MoveToNextFloor(eventTime)
	return fmt.Sprintf("%s Player [%d] went to the next floor", timeStr, playerID)
}

func (p *Processor) handlePrevFloor(timeStr string, playerID int, pl *player.Player) string {
	if !pl.InDungeon {
		return fmt.Sprintf("%s Player [%d] makes imposible move [5]", timeStr, playerID)
	}
	return fmt.Sprintf("%s Player [%d] makes imposible move [5]", timeStr, playerID)
}

func (p *Processor) handleEnterBoss(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [6]", timeStr, playerID)
	}
	if pl.CurrentFloor != p.cfg.Floors+1 {
		pl.EnterBossFloor(eventTime, p.cfg.Floors+1)
		return fmt.Sprintf("%s Player [%d] entered the boss's floor", timeStr, playerID)
	}
	return ""
}

func (p *Processor) handleKillBoss(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [7]", timeStr, playerID)
	}
	if pl.CurrentFloor != p.cfg.Floors+1 {
		return fmt.Sprintf("%s Player [%d] makes imposible move [7]", timeStr, playerID)
	}
	if pl.BossDefeated {
		return fmt.Sprintf("%s Player [%d] makes imposible move [7]", timeStr, playerID)
	}
	pl.KillBoss(eventTime)
	return fmt.Sprintf("%s Player [%d] killed the boss", timeStr, playerID)
}

func (p *Processor) handleLeave(timeStr string, playerID int, pl *player.Player, eventTime time.Time) string {
	if !pl.InDungeon && pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [8]", timeStr, playerID)
	}
	if pl.InDungeon {
		pl.LeaveDungeon(eventTime)
	}
	return fmt.Sprintf("%s Player [%d] left the dungeon", timeStr, playerID)
}

func (p *Processor) handleCannotContinue(timeStr string, playerID int, pl *player.Player, eventTime time.Time, reason string) string {
	if !pl.InDungeon {
		return fmt.Sprintf("%s Player [%d] makes imposible move [9]", timeStr, playerID)
	}
	pl.SetCannotContinue(eventTime)
	return fmt.Sprintf("%s Player [%d] cannot continue due to [%s]", timeStr, playerID, reason)
}

func (p *Processor) handleRestoreHealth(timeStr string, playerID int, pl *player.Player, healthStr string) (string, error) {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [10]", timeStr, playerID), nil
	}
	health, err := strconv.Atoi(healthStr)
	if err != nil || health < 0 {
		return "", fmt.Errorf("invalid health value: %s", healthStr)
	}
	pl.RestoreHealth(health)
	return fmt.Sprintf("%s Player [%d] has restored [%d] of health", timeStr, playerID, health), nil
}

func (p *Processor) handleDamage(timeStr string, playerID int, pl *player.Player, damageStr string) (string, error) {
	if !pl.InDungeon || !pl.IsActive() {
		return fmt.Sprintf("%s Player [%d] makes imposible move [11]", timeStr, playerID), nil
	}
	damage, err := strconv.Atoi(damageStr)
	if err != nil || damage < 0 {
		return "", fmt.Errorf("invalid damage value: %s", damageStr)
	}

	result := fmt.Sprintf("%s Player [%d] recieved [%d] of damage", timeStr, playerID, damage)
	isDead := pl.TakeDamage(damage)

	if isDead {
		result = result + "\n" + fmt.Sprintf("%s Player [%d] is dead", timeStr, playerID)
	}
	return result, nil
}

func (p *Processor) GetOutput() []string {
	return p.output
}

func (p *Processor) GetPlayers() map[int]*player.Player {
	return p.players
}

func ProcessEvents(events []string, cfg *config.Config, openTime, closeTime time.Time) (map[int]*player.Player, []string) {
	proc := NewProcessor(cfg, openTime, closeTime)

	for _, eventLine := range events {
		if err := proc.ProcessEvent(eventLine); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return proc.GetPlayers(), proc.GetOutput()
}
