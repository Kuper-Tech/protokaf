package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewConsumeCmd_NoTopicFlags(t *testing.T) {
	cmd := NewConsumeCmd()
	cmd.SetArgs([]string{"HelloRequest"})

	_, _, err := getCommandOut(t, cmd)

	require.Contains(t, err.Error(), `required flag(s) "group", "topic" not set`)
}

func Test_parseOffsetsFlag(t *testing.T) {
	t.Run("offset happy case", func(t *testing.T) {
		offset, err := parseOffsetFlag("1")
		require.Nil(t, err)
		require.EqualValues(t, 1, offset)
	})

	t.Run("error when offset is not a number", func(t *testing.T) {
		v, err := parseOffsetFlag("fdgdfg")
		require.ErrorIs(t, err, ErrInvalidOffset)
		require.EqualValues(t, -1, v)
	})

	t.Run("error empty offset", func(t *testing.T) {
		v, err := parseOffsetFlag("")
		require.ErrorIs(t, err, ErrOffsetNotSet)
		require.EqualValues(t, -1, v)
	})

	t.Run("error negative offset", func(t *testing.T) {
		v, err := parseOffsetFlag("-1")
		require.ErrorIs(t, err, ErrInvalidOffset)
		require.EqualValues(t, -1, v)
	})
}
