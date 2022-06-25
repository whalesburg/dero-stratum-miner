package console

import (
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

func New() (*readline.Instance, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[92mDERO Miner:\033[32m>>>\033[0m ",
		HistoryFile:     filepath.Join(os.TempDir(), "dero_miner_readline.tmp"),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("status"),
	readline.PcItem("version"),
	readline.PcItem("bye"),
	readline.PcItem("exit"),
	readline.PcItem("quit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
