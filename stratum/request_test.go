package stratum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var expr = `{"id":1,"jsonrpc":"2.0","method":"getwork","params":null}
` // the newline is important

func TestRequest(t *testing.T) {
	r := NewRequest(1, "getwork", nil)
	b, err := r.Parse()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expr, string(b))
}
