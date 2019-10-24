package packet_test

import (
	"testing"

	"github.com/aerogo/nano/packet"
	"github.com/akyoto/assert"
)

func TestPacket(t *testing.T) {
	p := packet.New(0, []byte("ping"))
	assert.Equal(t, p.Type(), packet.Type(0))
	assert.DeepEqual(t, p.Data(), []byte("ping"))
	assert.Equal(t, len(p), 1+len("ping"))
}
