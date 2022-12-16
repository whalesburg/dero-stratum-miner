//go:build !windows
// +build !windows

// Copyright 2017-2021 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package miner

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// we skip type as go will automatically identify type
const (
	UnixMax = 20
	OSXMax  = 20 // see this https://github.com/golang/go/issues/30401
)

type Limits struct {
	Current uint64
	Max     uint64
}

func init() {
	switch runtime.GOOS {
	case "darwin":
		unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Max: OSXMax, Cur: OSXMax}) // nolint: errcheck
	case "linux", "netbsd", "openbsd", "freebsd":
		unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Max: UnixMax, Cur: UnixMax}) // nolint: errcheck
	default: // nothing to do
	}
}

func Get() (*Limits, error) {
	var rLimit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		return nil, err
	}
	return &Limits{Current: uint64(rLimit.Cur), Max: uint64(rLimit.Max)}, nil //nolint: unconvert, otherwise bsd builds fail
}

/*
func Set(maxLimit uint64) error {
	rLimit := unix.Rlimit {Max:maxLimit, Cur:maxLimit}
	if runtime.GOOS == "darwin" && rLimit.Cur > OSXMax { //https://github.com/golang/go/issues/30401
		rLimit.Cur = OSXMax
	}
	return unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit)
}
*/
