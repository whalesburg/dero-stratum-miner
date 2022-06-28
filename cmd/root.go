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
	"time"

	"github.com/deroproject/derohe/rpc"
	"github.com/go-logr/logr"
	"github.com/muesli/coral"
	mcoral "github.com/muesli/mango-coral"
	"github.com/muesli/roff"
	"github.com/stratumfarm/dero-stratum-miner/internal/api"
	"github.com/stratumfarm/dero-stratum-miner/internal/config"
	"github.com/stratumfarm/dero-stratum-miner/internal/console"
	miner "github.com/stratumfarm/dero-stratum-miner/internal/dero-stratum-miner"
	"github.com/stratumfarm/dero-stratum-miner/internal/logging"
	"github.com/stratumfarm/dero-stratum-miner/internal/stratum"
	"github.com/stratumfarm/dero-stratum-miner/internal/version"
)

var cfg = config.NewEmpty()

var rootCmd = &coral.Command{
	Use:   "dero-stratum-miner",
	Short: "Dero Stratum Miner",
	RunE:  rootHandler,
}

func init() {
	rootCmd.AddCommand(versionCmd, manCmd)

	rootCmd.Flags().StringVarP(&cfg.Miner.Wallet, "wallet-address", "w", "", "wallet of the miner. Rewards will be sent to this address")
	rootCmd.MarkFlagRequired("wallet-address") // nolint: errcheck

	rootCmd.Flags().BoolVarP(&cfg.Miner.Testnet, "testnet", "t", false, "use testnet")
	rootCmd.Flags().StringVarP(&cfg.Miner.PoolURL, "daemon-rpc-address", "r", "pool.whalesburg.com:tbd", "stratum pool url")
	rootCmd.Flags().IntVarP(&cfg.Miner.Threads, "mining-threads", "m", runtime.GOMAXPROCS(0), "number of threads to use")

	rootCmd.Flags().BoolVar(&cfg.Logger.Debug, "debug", false, "enable debug mode")
	rootCmd.Flags().Int8Var(&cfg.Logger.CLogLevel, "console-log-level", 0, "console log level")
	rootCmd.Flags().Int8Var(&cfg.Logger.FLogLevel, "file-log-level", 0, "file log level")

	rootCmd.Flags().StringVar(&cfg.API.Listen, "api-listen", ":8080", "address to listen for API requests")
	rootCmd.Flags().BoolVar(&cfg.API.Disabled, "api-disabled", false, "disable the API server")
}

func Execute() error {
	return rootCmd.Execute()
}

func validateConfig(cfg *config.Config) error {
	if err := validateAddress(cfg.Miner.Testnet, cfg.Miner.Wallet); err != nil {
		return err
	}
	if cfg.Miner.Threads > runtime.GOMAXPROCS(0) {
		return fmt.Errorf("Mining threads is more than available CPUs. This is NOT optimal. Threads count: %d, max possible: %d", cfg.Miner.Threads, runtime.GOMAXPROCS(0))
	}

	return nil
}

func validateAddress(testnet bool, a string) error {
	addr, err := rpc.NewAddress(strings.Split(a, ".")[0])
	if err != nil {
		return err
	}

	if !addr.IsDERONetwork() {
		return fmt.Errorf("Invalid DERO address")
	}

	if !testnet != addr.IsMainnet() {
		if !testnet {
			return fmt.Errorf("Address belongs to DERO testnet and is invalid on current network")
		}
		return fmt.Errorf("Address belongs to DERO mainnet and is invalid on current network")
	}
	return nil
}

func rootHandler(cmd *coral.Command, args []string) error {
	if err := validateConfig(cfg); err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cli, err := console.New()
	if err != nil {
		log.Fatalln("failed to create console:", err)
	}

	exename, err := os.Executable()
	if err != nil {
		return err
	}
	f, err := os.Create(exename + ".log")
	if err != nil {
		return fmt.Errorf("Error while opening log file err: %s filename %s", err, exename+".log")
	}
	logger := logging.New(cli.Stdout(), f, cfg.Logger)

	ctx, cancel := context.WithCancel(cmd.Context())
	stc := newStratumClient(ctx, cfg.Miner.PoolURL, cfg.Miner.Wallet, logger)

	m, err := miner.New(ctx, cancel, cfg.Miner, stc, cli, logger)
	if err != nil {
		log.Fatalln(err)
	}
	defer m.Close()

	go func() {
		if err := m.Start(); err != nil {
			log.Fatalln(err)
		}
	}()

	if !cfg.API.Disabled {
		api := api.New(ctx, m, cfg.API, logger)
		defer api.Close()
		go func() {
			if err := api.Serve(); err != nil {
				log.Fatalln(err)
			}
		}()
	}

	select {
	case <-done:
	case <-ctx.Done():
	}
	cancel()

	return nil
}

func newStratumClient(ctx context.Context, url, addr string, logger logr.Logger) *stratum.Client {
	var useTLS bool
	if strings.HasPrefix(url, "stratum+tls://") || strings.HasPrefix(url, "stratum+ssl://") {
		useTLS = true
		url = strings.TrimPrefix(url, "stratum+tls://")
		url = strings.TrimPrefix(url, "stratum+ssl://")
	} else {
		useTLS = false
		url = strings.TrimPrefix(url, "stratum://")
		url = strings.TrimPrefix(url, "tcp://")
		url = strings.TrimPrefix(url, "stratum+tcp://")
	}
	opts := []stratum.Opts{
		stratum.WithUsername(addr),
		stratum.WithContext(ctx),
		stratum.WithReadTimeout(time.Second * 10),
		stratum.WithWriteTimeout(10 * time.Second),
		stratum.WithDebugLogger(func(s string) {
			logger.V(1).Info(s)
		}),
		stratum.WithErrorLogger(func(err error, s string) {
			logger.Error(err, s)
		}),
	}
	if useTLS {
		opts = append(opts, stratum.WithUseTLS())
	}

	return stratum.New(url, opts...)
}

var manCmd = &coral.Command{
	Use:                   "man",
	Short:                 "generates the manpages",
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	Args:                  coral.NoArgs,
	RunE: func(cmd *coral.Command, args []string) error {
		manPage, err := mcoral.NewManPage(1, rootCmd)
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
		return err
	},
}

var versionCmd = &coral.Command{
	Use:   "version",
	Short: "Print the version info",
	Run: func(cmd *coral.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.Commit)
		fmt.Printf("Date: %s\n", version.Date)
		fmt.Printf("Build by: %s\n", version.BuiltBy)
	},
}
