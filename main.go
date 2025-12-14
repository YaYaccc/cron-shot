package main

import (
	"cron-shot/gui"
	"cron-shot/logging"
)

func main() {
	defer logging.RecoverPanic("main")
	gui.Run()
}
