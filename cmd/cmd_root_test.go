package cmd

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewProduceCmd(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"produce",
		"HelloRequest",
		"--topic", "Test_NewProduceCmd",
		"--proto", "../internal/proto/testdata/example.proto",
		"--timeout", "5s",
		"-d", `{"name": "Alice", "age": 11}`,
	})

	stdout, stderr, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Contains(t, stdout, "Message produced")
	assert.Empty(t, stderr)
}

func Test_NewProduceCmd_WithTrace(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"produce",
		"HelloRequest",
		"--topic", "Test_NewProduceCmd",
		"--proto", "../internal/proto/testdata/example.proto",
		"--timeout", "5s",
		"--trace",
		"-d", `{"name": "Alice", "age": 11}`,
	})

	stdout, stderr, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Contains(t, stdout, "Message produced")
	assert.Empty(t, stderr)
}

func Test_NewProduceCmd_Random(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"produce",
		"HelloRequest",
		"--topic", "Test_NewProduceCmd",
		"--proto", "../internal/proto/testdata/example.proto",
		"--timeout", "5s",
		"--seed", "1",
		"--count", "3",
		"-d", `{"name": {{randomString 5 | quote}}, "age": {{randomNumber 1 10}}}`,
	})

	stdout, stderr, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Contains(t, stdout, "Producing 3 messages...")
	assert.Contains(t, stdout, "Prepared protobuf message: name:\"BpLnf\" age:7")
	assert.Empty(t, stderr)
}

func Test_NewConsumeCmd(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"consume",
		"HelloRequest",
		"--group", "test",
		"--topic", "Test_NewConsumeCmd",
		"--proto", "../internal/proto/testdata/example.proto",
		"--count", "1",
	})

	pr, pw := io.Pipe()
	cmd.SetOut(pw)

	setLogger(nopSync{pw}, "debug", "")
	go cmd.Execute() //nolint:errcheck

	r := bufio.NewReader(pr)
	t.Run("wait for consume", func(t *testing.T) {
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				t.Fatal(err)
			}

			if strings.Contains(string(line), "Consume topics") {
				break
			}
		}
	})
}

func Test_NewListCmd(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"list",
	})

	stdout, _, err := getCommandOut(t, cmd)

	assert.Nil(t, err)
	assert.Contains(t, stdout, "broker 1 \"127.0.0.1:9092\"")
	assert.Contains(t, stdout, "topic \"Test_NewProduceCmd\", partitions: 1")
}
