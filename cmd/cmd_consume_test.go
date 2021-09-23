package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewConsumeCmd_NoTopicFlags(t *testing.T) {
	cmd := NewProduceCmd()
	cmd.SetArgs([]string{"HelloRequest"})

	_, _, err := getCommandOut(t, cmd)

	assert.Contains(t, err.Error(), `required flag(s) "topic" not set`)
}
