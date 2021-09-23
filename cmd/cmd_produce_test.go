package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewProduceCmd_NoTopicFlags(t *testing.T) {
	cmd := NewProduceCmd()
	cmd.SetArgs([]string{"HelloRequest"})

	_, _, err := getCommandOut(t, cmd)

	assert.Contains(t, err.Error(), `required flag(s) "topic" not set`)
}

func Test_NewProduceCmd_TemplateFunctionsPrint(t *testing.T) {
	cmd := NewProduceCmd()
	cmd.SetArgs([]string{"--template-functions-print"})

	stdout, stderr, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Empty(t, stderr)

	assert.Contains(t, stdout, `Template functions:`)
	assert.Contains(t, stdout, `randomString`)
}

func Test_NewProduceCmd_JaegerConfigPrint(t *testing.T) {
	cmd := NewProduceCmd()
	cmd.SetArgs([]string{"--jaeger-config-print"})

	stdout, stderr, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Empty(t, stderr)

	assert.Contains(t, stdout, `Jaeger config`)
	assert.Contains(t, stdout, `"LocalAgentHostPort": "0.0.0.0:6831"`)
}
