package reporter

import (
	"fmt"
	"time"

	"dungeon/internal/config"
	"dungeon/internal/player"
)

func PrintOutput(output []string) {
	fmt.Println("\n=== OUTPUT ===")
	for _, line := range output {
		fmt.Println(line)
	}
}

func PrintFinalReport(players map[int]*player.Player, cfg *config.Config) {
	fmt.Println("\nFinal report:")

	for _, pl := range players {
		status := determineStatus(pl, cfg)
		totalTime := pl.GetTotalTime()
		avgClearTime := pl.GetAverageClearTime(cfg.Floors)
		bossTime := pl.BossKillDuration
		hp := pl.Health

		if status == "DISQUAL" && (pl.Disqualified || !pl.Registered) {
			hp = 100
			totalTime = 0
		}

		formatDur := func(d time.Duration) string {
			if d < 0 {
				d = 0
			}
			h := int(d.Hours())
			m := int(d.Minutes()) % 60
			s := int(d.Seconds()) % 60
			return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
		}

		fmt.Printf("[%s] %d [%s, %s, %s] HP:%d\n",
			status, pl.ID,
			formatDur(totalTime),
			formatDur(avgClearTime),
			formatDur(bossTime),
			hp)
	}
}

func determineStatus(pl *player.Player, cfg *config.Config) string {
	// SUCCESS: босс побеждён И игрок вышел из подземелья
	if pl.BossDefeated && !pl.InDungeon {
		return "SUCCESS"
	}
	if pl.Dead {
		return "FAIL"
	}
	if pl.Disqualified || !pl.Registered {
		return "DISQUAL"
	}
	if pl.CannotContinue {
		return "DISQUAL"
	}
	return "FAIL"
}
