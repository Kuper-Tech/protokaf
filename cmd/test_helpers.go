package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func getCommandOut(t *testing.T, cmd *cobra.Command) (stdout string, stderr string, errCmd error) {
	o := new(bytes.Buffer)
	e := new(bytes.Buffer)

	cmd.SetOut(o)
	cmd.SetErr(e)

	setLogger(nopSync{o}, "debug", "")
	errCmd = cmd.Execute()

	return o.String(), e.String(), errCmd
}
