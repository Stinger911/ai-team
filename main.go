package main

import (
	"ai-team/cmd"
	"ai-team/pkg/logger"
	"log"
	"os"
)

func main() {
	logger.SetLogLevelFromEnv()
	// Startup check: warn if config.yaml is missing
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		log.Printf("WARNING: config.yaml not found in current directory. The application may not function correctly.")
	}
	cmd.ExecuteCmd()
}
