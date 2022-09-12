//go:build linux && (arm || arm64)
// +build linux
// +build arm arm64

package cmd

import (
	"time"

	"github.com/deroproject/derohe/astrobwt/astrobwtv3"
)

// this adds some delay to the startup of the miner.
// while waiting, it generates high CPU load. Otherwise not all cores are detected on android (sometimes (it's weird)).

func bootDelay() {
	x := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
	t := time.NewTimer(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			return
		default:
			go astrobwtv3.AstroBWTv3(x)
		}
	}
}
