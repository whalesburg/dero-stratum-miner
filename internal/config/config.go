package config

type Config struct {
	Miner  *Miner
	Logger *Logger
	API    *API
}

type Miner struct {
	Wallet  string
	Testnet bool
	PoolURL string
	Threads int
}

type Logger struct {
	Debug     bool
	CLogLevel int8
	FLogLevel int8
}

type API struct {
	Transport string
	Listen    string
	Enabled   bool
}

// NewEmpty returns a new empty config
func NewEmpty() *Config {
	return &Config{
		Miner:  &Miner{},
		Logger: &Logger{},
		API:    &API{},
	}
}
