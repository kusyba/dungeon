package reader

import (
	"bufio"
	"io"
	"os"
)

func ReadEvents(stdin io.Reader, fallbackFile string) []string {
	scanner := bufio.NewScanner(stdin)
	events := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		events = append(events, line)
	}

	if len(events) == 0 {
		file, err := os.Open(fallbackFile)
		if err == nil {
			defer file.Close()
			scanner = bufio.NewScanner(file)
			for scanner.Scan() {
				if line := scanner.Text(); line != "" {
					events = append(events, line)
				}
			}
		}
	}

	return events
}
