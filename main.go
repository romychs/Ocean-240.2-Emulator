package main

import (
	"z80em/config"
	"z80em/logger"
	"z80em/okean240"
)

var Version = "v1.0.0"
var BuildTime = "2026-03-01"

func main() {

	// base log init
	logger.InitLogging()

	// load config yml file
	config.LoadConfig()

	conf := config.GetConfig()

	// Reconfigure logging by config values
	logger.ReconfigureLogging(conf)

	println("Init computer")
	computer := okean240.New(config)
	println("Run computer")
	computer.Run()
}
