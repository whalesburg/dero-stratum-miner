# dero-stratum-miner
[![lint](https://github.com/stratumfarm/dero-stratum-miner/actions/workflows/lint.yml/badge.svg)](https://github.com/stratumfarm/dero-stratum-miner/actions/workflows/lint.yml)
[![goreleaser](https://github.com/stratumfarm/dero-stratum-miner/actions/workflows/release.yml/badge.svg)](https://github.com/stratumfarm/dero-stratum-miner/actions/workflows/release.yml)
[![Discord](https://img.shields.io/discord/955758990682390568?logo=discord&logoColor=white&labelColor=5865F2&color=gray)](https://discord.gg/GSacSHyEBP)
[![telegram chat](https://img.shields.io/badge/telegram-chat-gray?labelColor=0088cc)](https://t.me/+KmaphwptVMQ2ZDBk)


## 💡 About
dero-stratum-miner adds support for mining to a stratum pool using the official [derod](https://github.com/deroproject/derohe) mining algorithm.


## 📦 Installation
The latest release can always be found [here](https://github.com/stratumfarm/dero-stratum-miner/releases).


### Linux

#### Manual installation
1. Open a terminal
2. Download the latest release  
`curl -sLJO "https://github.com/stratumfarm/dero-stratum-miner/releases/download/v1.0.0/dero-stratum-miner-v1.0.0-linux-amd64.tar.gz"`

3. Unpack the archive  
`tar xavf dero-stratum-miner-v1.0.0-linux-amd64.tar.gz`

4. Make the file executable  
`cd dero-stratum-miner-v1.0.0-linux-amd64 && chmod u+x dero-stratum-miner`

The miner can be started by using the command `./dero-stratum-miner`


#### Using a package manager
```bash
# -> download the latest release file first.

# debian / ubuntu
dpkg -i dero-stratum-miner-v1.0.0-linux-amd64.deb

# rhel / fedora / suse
rpm -i dero-stratum-miner-v1.0.0-linux-amd64.rpm

# alpine
apk add --allow-untrusted dero-stratum-miner-v1.0.0-linux-amd64.apk
```


#### Arch (btw)
```
$ yay -S dero-stratum-miner-bin
```


### Windows
1. Download the latest release for windows
2. Unzip the archive
3. Open a terminal in the newly created folder with `right click -> "Open in Windows Terminal"`

The miner can be started by using the command `.\dero-stratum-miner.exe`


### MacOS
1. Download the latest release for windows
2. Unzip the archive
3. Open a terminal in the newly created folder with `right click -> "Open Terminal"`

The miner can be started by using the command `./dero-stratum-miner`


### Android
1. Install termux (
    [Playstore](https://play.google.com/store/apps/details?id=com.termux&gl=US),
    [F-Droid](https://f-droid.org/en/packages/com.termux/)
)
2. Open termux
3. Update all packages  
`pkg update`

4. Download the latest release  
`curl -sLJO "https://github.com/stratumfarm/dero-stratum-miner/releases/download/v1.0.0/dero-stratum-miner-v1.0.0-linux-armv7.tar.gz"`

5. Unpack the archive  
`tar xavf dero-stratum-miner-v1.0.0-linux-amd64.tar.gz`

6. Make the file executable  
`cd dero-stratum-miner-v1.0.0-linux-armv7 && chmod u+x dero-stratum-miner`

The miner can be started by using the command `./dero-stratum-miner`


## 🚀 Usage

### Start the miner
To simply start the miner and get going, you can use the following command:
```
$ ./dero-stratum-miner -w $YOUR_WALLET
```

### Disable TLS
By default the dero-stratum-miner tries to connect over a tls encrypted port. This can be disabled using the following command:
```
$ ./dero-stratum-miner -w $YOUR_WALLET -r stratum://pool.whalesburg.com:tbd
```

### Enabled the api
To fetch stats from the miner, an internal API can be enabled by using the `--api-enabled` parameter.
```
$ ./dero-stratum-miner -w $YOUR_WALLET --api-enabled
```

### Full Help
```
$ ./dero-stratum-miner help
Dero Stratum Miner

Usage:
  dero-stratum-miner [flags]
  dero-stratum-miner [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version info

Flags:
      --api-enabled                 enable the API server
      --api-listen string           address to listen for API requests (default ":8080")
      --console-log-level int8      console log level
  -r, --daemon-rpc-address string   stratum pool url (default "pool.whalesburg.com:tbd")
      --debug                       enable debug mode
      --file-log-level int8         file log level
  -h, --help                        help for dero-stratum-miner
  -m, --mining-threads int          number of threads to use (default 32)
  -t, --testnet                     use testnet
  -w, --wallet-address string       wallet of the miner. Rewards will be sent to this address

Use "dero-stratum-miner [command] --help" for more information about a command.
```


## 🆘 Support 
If you want to report a bug, please [open an issue](https://github.com/stratumfarm/dero-stratum-miner/issues/new/choose) on github.  
For general support please join our telegram group or discord server.
- [Discord](https://discord.gg/GSacSHyEBP)
- [Telegram](https://t.me/+KmaphwptVMQ2ZDBk)


## 📜 Licensing
_TODO_