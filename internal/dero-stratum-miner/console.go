package miner

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/stratumfarm/dero-stratum-miner/internal/version"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, "\t\033[1mhelp\033[0m\t\tthis help\n")
	io.WriteString(w, "\t\033[1mstatus\033[0m\t\tShow general information\n")
	io.WriteString(w, "\t\033[1mbye\033[0m\t\tQuit the miner\n")
	io.WriteString(w, "\t\033[1mversion\033[0m\t\tShow version\n")
	io.WriteString(w, "\t\033[1mexit\033[0m\t\tQuit the miner\n")
	io.WriteString(w, "\t\033[1mquit\033[0m\t\tQuit the miner\n")
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

func (c *Client) refreshConsole() {
	last_counter := uint64(0)
	last_counter_time := time.Now()

	mining := true
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// only update prompt if needed
		if last_counter != c.counter {
			// choose color based on urgency
			color := "\033[33m" // default is green color

			mining_string := ""

			if mining {
				mining_speed := float64(c.counter-last_counter) / (float64(uint64(time.Since(last_counter_time))) / 1000000000.0)
				c.hashrate = uint64(mining_speed)
				last_counter = c.counter
				last_counter_time = time.Now()
				switch {
				case mining_speed > 1000000:
					mining_string = fmt.Sprintf("MINING @ %.3f MH/s", float32(mining_speed)/1000000.0)
				case mining_speed > 1000:
					mining_string = fmt.Sprintf("MINING @ %.3f KH/s", float32(mining_speed)/1000.0)
				case mining_speed > 0:
					mining_string = fmt.Sprintf("MINING @ %.0f H/s", mining_speed)
				}
			}

			testnet_string := ""
			if c.config.Testnet {
				testnet_string = "\033[31m TESTNET"
			}

			c.console.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO Miner: \033[0m"+color+"Shares %d Rejected %d \033[32m%s>%s>>\033[0m ", c.shareCounter, c.rejectedCounter, mining_string, testnet_string))
			c.console.Refresh()
		}
		time.Sleep(1 * time.Second)
	}
}
