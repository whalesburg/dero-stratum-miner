//go:build !(linux && (arm || arm64))
// +build !linux !arm,!arm64

package cmd

func bootDelay() {
}
