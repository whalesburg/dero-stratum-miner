//go:build !(linux && (arm || arm64))
// +build !linux !arm,!arm64

package dns

func BootstrapDNS(_ string) bool {
	return false
}
