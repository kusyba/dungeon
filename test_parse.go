package main

import (
	"fmt"
	"strings"
)

func main() {
	lines := []string{
		"[14:00:00] 1 1",
		"[14:00:00] 2 1",
		"[14:00:00] 1 2",
		"[14:10:00] 2 2",
	}
	
	for _, line := range lines {
		parts := strings.Fields(line)
		fmt.Printf("line: %s\n", line)
		fmt.Printf("  parts[0]=%q (time)\n", parts[0])
		fmt.Printf("  parts[1]=%q (should be eventID)\n", parts[1])
		fmt.Printf("  parts[2]=%q (should be playerID)\n", parts[2])
		fmt.Println()
	}
}
