package miner

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/whalesburg/dero-stratum-miner/internal/version"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")                               // nolint: errcheck
	io.WriteString(w, "\t\033[1mhelp\033[0m\t\tthis help\n")       // nolint: errcheck
	io.WriteString(w, "\t\033[1mbye\033[0m\t\tQuit the miner\n")   // nolint: errcheck
	io.WriteString(w, "\t\033[1mversion\033[0m\t\tShow version\n") // nolint: errcheck
	io.WriteString(w, "\t\033[1mexit\033[0m\t\tQuit the miner\n")  // nolint: errcheck
	io.WriteString(w, "\t\033[1mquit\033[0m\t\tQuit the miner\n")  // nolint: errcheck
}

func (c *Client) startConsole() {
	for {
		line, err := c.console.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				fmt.Print("Ctrl-C received, Exit in progress\n")
				c.cancel()
				os.Exit(0)
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			<-c.ctx.Done()
			break
		}

		line = strings.TrimSpace(line)
		lineParts := strings.Fields(line)

		command := ""
		if len(lineParts) >= 1 {
			command = strings.ToLower(lineParts[0])
		}

		switch {
		case line == "help":
			usage(c.console.Stderr())

		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				fmt.Println("say what?")
				break
			}
		case command == "version":
			fmt.Printf("Version %s OS:%s ARCH:%s \n", version.Version, runtime.GOOS, runtime.GOARCH)

		case strings.ToLower(line) == "bye":
			fallthrough
		case strings.ToLower(line) == "exit":
			fallthrough
		case strings.ToLower(line) == "quit":
			c.cancel()
			os.Exit(0)
		case line == "":
		default:
			fmt.Println("you said:", strconv.Quote(line))
		}
	}
}

func (c *Client) setPrompt(heightString, diffString, miningString, testnetString string) {
	if c.console == nil {
		return
	}
	c.console.SetPrompt(fmt.Sprintf("\033[1m\033[32mDero-Stratum-Miner: \033[0m%s %s \033[33mShares %d Rejected %d \033[32m%s>%s>>\033[0m ", heightString, diffString, c.GetTotalShares(), c.GetRejectedShares(), miningString, testnetString))
	c.console.Refresh()
}
