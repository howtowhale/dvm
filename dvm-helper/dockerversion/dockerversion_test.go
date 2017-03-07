package dockerversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripLeadingV(t *testing.T) {
	v := Parse("v1.0.0")
	assert.Equal(t, "1.0.0", v.String())
}
