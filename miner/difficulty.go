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

// ripoff from blockchain folder
import (
	"math/big"

	"github.com/deroproject/derohe/cryptography/crypto"
)

var (
	// bigZero is 0 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigZero = big.NewInt(0)

	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)

	// enabling this will simulation mode with hard coded difficulty set to 1
	// the variable is knowingly not exported, so no one can tinker with it
	//simulation = false // simulation mode is disabled
)

// HashToBig converts a PoW has into a big.Int that can be used to
// perform math comparisons.
func HashToBig(buf crypto.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	blen := len(buf) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

func ConvertIntegerDifficultyToBig(difficultyi *big.Int) *big.Int {
	if difficultyi.Cmp(bigZero) == 0 { // if work_pow is less than difficulty
		panic("difficulty can never be zero")
	}

	return new(big.Int).Div(oneLsh256, difficultyi)
}

// this function check whether the pow hash meets difficulty criteria
// however, it take diff in  bigint format
func CheckPowHashBig(powHash crypto.Hash, bigDifficultyInteger *big.Int) bool {
	bigPowHash := HashToBig(powHash)

	bigDifficulty := ConvertIntegerDifficultyToBig(bigDifficultyInteger)

	return bigPowHash.Cmp(bigDifficulty) <= 0 // if work_pow is less than difficulty
}
