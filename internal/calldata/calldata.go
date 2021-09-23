package calldata

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/google/uuid"
)

var (
	tmpl *template.Template
)

func init() {
	fnMap := make(template.FuncMap, len(funcs))
	for _, fn := range funcs {
		fnMap[fn.Name] = fn.Func
	}

	tmpl = template.New("calldata").Funcs(fnMap)
}

type Var struct {
	Name string
	Desc string
}

var (
	variables = []Var{
		{Name: "RequestNumber", Desc: "Request number for data"},
		{Name: "Timestamp", Desc: "Timestamp in RFC3339 format"},
		{Name: "TimestampUnix", Desc: "Timestamp as unix time in seconds"},
		{Name: "TimestampUnixMilli", Desc: "Timestamp as unix time in milliseconds"},
		{Name: "TimestampUnixNano", Desc: "Timestamp as unix time in nanoseconds"},
		{Name: "UUID", Desc: "Generated UUIDv4"},
	}
)

type CallData struct {
	RequestNumber      int64
	Timestamp          string
	TimestampUnix      int64
	TimestampUnixMilli int64
	TimestampUnixNano  int64
	UUID               string
}

func ParseTemplate(data []byte) (*template.Template, error) {
	t, err := tmpl.Parse(string(data))
	if err != nil {
		return nil, err
	}

	return t, nil
}

// PrintFuncs prints list of functions to output.
func PrintFuncs(output io.Writer) {
	w := tabwriter.NewWriter(output, 0, 0, 4, ' ', 0)

	fmt.Fprintln(w, "Template variables:\t")
	for _, vr := range variables {
		fmt.Fprintf(w, "  .%s\t%s\n", vr.Name, vr.Desc)
	}

	fmt.Fprintln(w, "\t")
	fmt.Fprintln(w, "Template functions:\t")
	for _, fn := range funcs {
		fmt.Fprintf(w, "  %s\t%s\n", fn.Name, fn.Desc)
	}
	w.Flush()
}

func NewCallData(reqNum int) *CallData {
	now := time.Now()
	nowNano := now.UnixNano()

	return &CallData{
		RequestNumber:      int64(reqNum),
		Timestamp:          now.Format(time.RFC3339),
		TimestampUnix:      now.Unix(),
		TimestampUnixMilli: nowNano / 1e6,
		TimestampUnixNano:  nowNano,
		UUID:               uuid.NewString(),
	}
}

func (c *CallData) Execute(tmpl *template.Template) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)
	if err := tmpl.Execute(b, c); err != nil {
		return nil, err
	}

	return b, nil
}
