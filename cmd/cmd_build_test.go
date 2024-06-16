package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewBuildCmd(t *testing.T) {
	cmd := NewBuildCmd()
	NewFlags(cmd).Init()
	cmd.SetArgs([]string{"HelloRequest", "--proto", "../internal/proto/testdata/example.proto"})

	expected := `{
  "name": "",
  "age": 0,
  "amount": 0,
  "status": "PENDING",
  "numbers": [
    {
      "num": 0
    }
  ],
  "data": {
    "timestamp": "1970-01-01T00:00:00Z",
    "completed": false,
    "properties": {
      "0": ""
    },
    "blob": ""
  }
}`

	stdout, stderr, err := getCommandOut(t, cmd)

	require.Nil(t, err)
	require.Equal(t, "", stderr)
	require.Contains(t, stdout, expected)
}
