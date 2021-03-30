package main

import (
	sb "github.com/Petrify/hskl-bot/school-discord"
	"github.com/Petrify/simp-core"
)

func main() {
	sb.Start()
	simp.Wait()
}
