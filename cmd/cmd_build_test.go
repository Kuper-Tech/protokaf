package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewBuildCmd(t *testing.T) {
	cmd := NewBuildCmd()
	NewFlags(cmd).Init()
	cmd.SetArgs([]string{"ExampleMessage", "--proto", "../internal/proto/testdata/types.proto"})

	expected := `{
  "int32Field": 0,
  "int64Field": "0",
  "uint32Field": 0,
  "uint64Field": "0",
  "sint32Field": 0,
  "sint64Field": "0",
  "fixed32Field": 0,
  "fixed64Field": "0",
  "sfixed32Field": 0,
  "sfixed64Field": "0",
  "floatField": 0,
  "doubleField": 0,
  "boolField": true,
  "stringField": "",
  "bytesField": "",
  "enumField": "UNKNOWN",
  "messageField": {
    "nestedInt32": 0,
    "nestedString": ""
  },
  "repeatedInt32Field": [
    0
  ],
  "repeatedStringField": [
    ""
  ],
  "mapStringInt32Field": {
    "": 0
  },
  "mapInt32MessageField": {
    "0": {
      "nestedInt32": 0,
      "nestedString": ""
    }
  },
  "option1": 0,
  "anyField": null,
  "timestampField": "1970-01-01T00:00:00Z",
  "durationField": "0s",
  "structField": {
    "": null
  },
  "valueField": null,
  "listValueField": [
    null
  ],
  "boolValueField": true,
  "bytesValueField": null,
  "doubleValueField": 0,
  "floatValueField": 0,
  "int32ValueField": 0,
  "int64ValueField": "0",
  "stringValueField": "",
  "uint32ValueField": 0,
  "uint64ValueField": "0"
}`

	stdout, stderr, err := getCommandOut(t, cmd)

	require.Nil(t, err)
	require.Equal(t, "", stderr)
	require.Contains(t, stdout, expected)
}
