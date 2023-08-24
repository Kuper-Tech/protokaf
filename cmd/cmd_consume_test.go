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
	t.Run("offsets for multiple topics", func(t *testing.T) {
		offsets, err := parseOffsetsFlag([]string{"topic1:0", "topic2:1", "topic3:234", "topic4:123"})
		require.Nil(t, err)
		require.Equal(t, map[string]int64{
			"topic1": 0,
			"topic2": 1,
			"topic3": 234,
			"topic4": 123,
		}, offsets)
	})

	t.Run("global offset", func(t *testing.T) {
		offsets, err := parseOffsetsFlag([]string{"123"})
		require.Nil(t, err)
		require.Equal(t, map[string]int64{
			globalOffset: 123,
		}, offsets)
	})

	t.Run("global offset", func(t *testing.T) {
		offsets, err := parseOffsetsFlag([]string{"123"})
		require.Nil(t, err)
		require.Equal(t, map[string]int64{
			globalOffset: 123,
		}, offsets)
	})

	t.Run("global offset has priority over other definitions", func(t *testing.T) {
		offsets, err := parseOffsetsFlag([]string{"123", "topic1:123", "topic2:321"})
		require.Nil(t, err)
		require.Equal(t, map[string]int64{
			globalOffset: 123,
		}, offsets)
	})

	t.Run("error when offset is not a number", func(t *testing.T) {
		v, err := parseOffsetsFlag([]string{"topic1:123", "topic2:abc"})
		require.Nil(t, v)
		require.ErrorIs(t, err, ErrInvalidOffset)
	})

	t.Run("skip empty offset", func(t *testing.T) {
		v, err := parseOffsetsFlag([]string{""})
		require.Equal(t, map[string]int64{}, v)
		require.NoError(t, err)

		v, err = parseOffsetsFlag([]string{"", "topic1:123", ""})
		require.Equal(t, map[string]int64{"topic1": 123}, v)
		require.NoError(t, err)
	})
}
