package main

import (
	"ai-team/cmd"
	"ai-team/pkg/logger"
)

func main() {
	logger.SetLogLevelFromEnv()
	cmd.ExecuteCmd()
}
