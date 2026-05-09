package test

import (
	"strings"
	"testing"
	"time"

	"dungeon/internal/config"
	"dungeon/internal/player"
	"dungeon/internal/processor"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{"Valid config", config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2}, false},
		{"Invalid floors", config.Config{Floors: 0, Monsters: 2, OpenAt: "14:05:00", Duration: 2}, true},
		{"Invalid open time", config.Config{Floors: 2, Monsters: 2, OpenAt: "25:05:00", Duration: 2}, true},
		{"Negative duration", config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlayerRegistration(t *testing.T) {
	pl := player.NewPlayer(1, 2)

	if pl.Registered {
		t.Error("New player should not be registered")
	}

	pl.Register()
	if !pl.Registered {
		t.Error("Player should be registered after Register()")
	}

	result := pl.Register()
	if result {
		t.Error("Second registration should return false")
	}
}

func TestPlayerHealth(t *testing.T) {
	pl := player.NewPlayer(1, 2)

	if pl.Health != 100 {
		t.Errorf("Initial health should be 100, got %d", pl.Health)
	}

	pl.RestoreHealth(30)
	if pl.Health != 100 {
		t.Errorf("Health should cap at 100, got %d", pl.Health)
	}

	pl.Health = 50
	pl.TakeDamage(30)
	if pl.Health != 20 {
		t.Errorf("Health should be 20 after 30 damage, got %d", pl.Health)
	}

	isDead := pl.TakeDamage(30)
	if !isDead {
		t.Error("Player should be dead after taking 30 damage when health is 20")
	}
	if pl.Health != 0 {
		t.Errorf("Health should be 0 after death, got %d", pl.Health)
	}
}

func TestKillMonster(t *testing.T) {
	pl := player.NewPlayer(1, 2)
	enterTime, _ := time.Parse("15:04:05", "14:40:00")
	pl.EnterDungeon(enterTime)

	eventTime, _ := time.Parse("15:04:05", "14:41:00")
	cleared := pl.KillMonster(eventTime, 2)

	if cleared {
		t.Error("Floor should not be cleared after first monster")
	}
	if pl.MonstersKilled != 1 {
		t.Errorf("Monsters killed should be 1, got %d", pl.MonstersKilled)
	}

	eventTime2, _ := time.Parse("15:04:05", "14:45:00")
	cleared = pl.KillMonster(eventTime2, 2)

	if !cleared {
		t.Error("Floor should be cleared after second monster")
	}
	if !pl.FloorsCompleted[1] {
		t.Error("Floor 1 should be marked as completed")
	}
}

func TestImpossibleMove(t *testing.T) {
	cfg, _ := config.Load("../config.json")
	openTime, _ := cfg.GetOpenTime()
	closeTime := cfg.GetCloseTime(openTime)

	// Время без квадратных скобок
	events := []string{
		"14:00:00 1 1",
		"14:40:00 2 1",
		"14:41:00 5 1",
	}

	_, output := processor.ProcessEvents(events, cfg, openTime, closeTime)

	found := false
	for _, line := range output {
		if strings.Contains(line, "makes imposible move") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected impossible move message for event 5")
	}
}

func TestDungeonSuccess(t *testing.T) {
	cfg, _ := config.Load("../config.json")
	cfg.Floors = 1
	cfg.Monsters = 1
	openTime, _ := cfg.GetOpenTime()
	closeTime := cfg.GetCloseTime(openTime)

	// Время без квадратных скобок
	events := []string{
		"14:00:00 1 1",
		"14:40:00 2 1",
		"14:41:00 3 1",
		"14:42:00 4 1",
		"14:43:00 6 1",
		"14:44:00 7 1",
		"14:45:00 8 1",
	}

	players, _ := processor.ProcessEvents(events, cfg, openTime, closeTime)

	pl := players[1]
	if pl == nil {
		t.Error("Player should exist")
		return
	}
	if !pl.IsSuccess(cfg.Floors) {
		t.Error("Player should have succeeded")
	}
	if !pl.BossDefeated {
		t.Error("Boss should be defeated")
	}
}

func TestDeath(t *testing.T) {
	cfg, _ := config.Load("../config.json")
	openTime, _ := cfg.GetOpenTime()
	closeTime := cfg.GetCloseTime(openTime)

	// Время без квадратных скобок
	events := []string{
		"14:00:00 1 1",
		"14:40:00 2 1",
		"14:41:00 11 1 100",
	}

	players, output := processor.ProcessEvents(events, cfg, openTime, closeTime)

	pl := players[1]
	if pl == nil {
		t.Error("Player should exist")
		return
	}
	if !pl.Dead {
		t.Error("Player should be dead")
	}

	found := false
	for _, line := range output {
		if strings.Contains(line, "is dead") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected death message")
	}
}
