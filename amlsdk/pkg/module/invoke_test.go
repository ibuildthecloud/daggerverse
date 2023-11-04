package module

import (
	"testing"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/pkg/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	var val value.Value
	err := aml.Unmarshal([]byte("\"yo\""), &val)
	require.NoError(t, err)
	assert.Equal(t, "yo", string(val.(value.String)))
}
