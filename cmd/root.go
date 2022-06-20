package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/deroproject/derohe/rpc"
	"github.com/muesli/coral"
	miner "github.com/stratumfarm/dero-stratum-miner/internal/dero-stratum-miner"
	"github.com/stratumfarm/dero-stratum-miner/internal/stratum"
)

var cfg = &miner.Config{}

var rootCmd = &coral.Command{
	Use:   "dero-stratum-miner",
	Short: "Dero Stratum Miner",
	RunE:  rootHandler,
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.Wallet, "wallet-address", "w", "", "wallet of the miner. Rewards will be sent to this address")
	rootCmd.MarkFlagRequired("wallet-address")

	rootCmd.Flags().BoolVarP(&cfg.Testnet, "testnet", "t", false, "use testnet")
	rootCmd.Flags().StringVarP(&cfg.PoolURL, "daemon-rpc-address", "r", "pool.whalesburg.com:tbd", "stratum pool url")
	rootCmd.Flags().IntVarP(&cfg.Threads, "mining-threads", "m", runtime.GOMAXPROCS(0), "number of threads to use")

	rootCmd.Flags().BoolVar(&cfg.Debug, "debug", false, "enable debug mode")
	rootCmd.Flags().Int8Var(&cfg.CLogLevel, "console-log-level", 0, "console log level")
	rootCmd.Flags().Int8Var(&cfg.FLogLevel, "file-log-level", 0, "file log level")

}

func Execute() error {
	return rootCmd.Execute()
}

func validateConfig(cfg *miner.Config) error {
	if err := validateAddress(cfg.Testnet, cfg.Wallet); err != nil {
		return err
	}
	if cfg.Threads > runtime.GOMAXPROCS(0) {
		return fmt.Errorf("Mining threads is more than available CPUs. This is NOT optimal. Threads count: %d, max possible: %d", cfg.Threads, runtime.GOMAXPROCS(0))
	}

	return nil
}

func validateAddress(testnet bool, a string) error {
	addr, err := rpc.NewAddress(a)
	if err != nil {
		return err
	}

	if !addr.IsDERONetwork() {
		return fmt.Errorf("Invalid DERO address")
	}

	if !testnet != addr.IsMainnet() {
		if !testnet {
			return fmt.Errorf("Address belongs to DERO testnet and is invalid on current network")
		} else {
			return fmt.Errorf("Address belongs to DERO mainnet and is invalid on current network")
		}
	}
	return nil
}

func rootHandler(cmd *coral.Command, args []string) error {
	if err := validateConfig(cfg); err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(cmd.Context())
	m, err := miner.New(ctx, cancel, newStratumClient(ctx, cfg.PoolURL, cfg.Wallet), cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer m.Close()

	go func() {
		if err := m.Start(); err != nil {
			log.Fatalln(err)
		}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	cancel()

	return nil
}

func newStratumClient(ctx context.Context, url, addr string) *stratum.Client {
	var useTLS bool
	if strings.HasPrefix(url, "tls://") || strings.HasPrefix(url, "ssl://") || strings.HasPrefix(url, "https://") {
		useTLS = true
		url = strings.TrimPrefix(url, "tls://")
		url = strings.TrimPrefix(url, "ssl://")
		url = strings.TrimPrefix(url, "https://")
	} else {
		useTLS = false
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "tcp://")
	}
	opts := []stratum.Opts{
		stratum.WithUsername(addr),
		stratum.WithContext(ctx),
	}
	if useTLS {
		opts = append(opts, stratum.WithUseTLS())
	}
	return stratum.New(url, opts...)
}
