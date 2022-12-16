package main

import (
	"os"

	"github.com/whalesburg/dero-stratum-miner/cmd"
)

func main() {
	/* client := stratum.New("pool.whalesburg.com:4300",
		stratum.WithAgentName("wheeeee"),
		stratum.WithUsername("dero1qy2vshednrtqgtfkuagzne0xl0dcnxcq4tscg3qz3keekq9vsmtqcqg8gvm0e"),
		stratum.WithReadTimeout(time.Second*5),
		stratum.WithWriteTimeout(5*time.Second),
	)
	if err := client.Dial(context.Background()); err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 10) */
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
