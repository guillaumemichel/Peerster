package gossiper

import "strconv"

// LogCode associated with different flags
type LogCode int

const (
	logHW1 LogCode = 0
	logHW2 LogCode = 1
	logHW3 LogCode = 2
)

// ShouldPrint return true if it should print log
func (g *Gossiper) ShouldPrint(code LogCode, lvl int) bool {
	i, err := strconv.Atoi(string([]rune(g.LogLvl)[code]))
	if err != nil {
		g.Printer.Println(err)
	}
	if i < lvl {
		return false
	}
	return true
}
