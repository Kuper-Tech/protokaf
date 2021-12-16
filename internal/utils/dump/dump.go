package dump

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Logger interface {
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
	Debugf(string, ...interface{})
	Debug(...interface{})
}

// Pair is a name, value pair.
type Pair struct {
	Name  string
	Value interface{}
}

type Pairs []Pair

// Dump dumps list of pairs with Logger.
func (p Pairs) Dump(log Logger) {
	maxLen := 0
	for _, m := range p {
		if n := len(m.Name); maxLen < n {
			maxLen = n
		}
	}

	nameWithPad := func(n string) string {
		pad := strings.Repeat(" ", maxLen-len(n)+1)
		return n + pad
	}

	log.Debug(titleStd("Dump begin"))

	for _, m := range p {
		n, v := m.Name, m.Value

		switch x := v.(type) {
		case []byte:
			log.Debugf("%s: <hex dump>\n%s", nameWithPad(n), hex.Dump(x))

		default:
			if v == "" {
				v = "(empty)"
			}

			log.Debugf("%s: %v", nameWithPad(n), v)
		}
	}
}

func marshalJSONCustom(msg *dynamic.Message) func() ([]byte, error) {
	return func() ([]byte, error) {
		return msg.MarshalJSONPB(&jsonpb.Marshaler{Indent: "  ", EmitDefaults: true})
	}
}

// DynamicMessage dumps dynamic.Message.
func DynamicMessage(log Logger, name, output string, msg *dynamic.Message) {
	marshaller := marshalJSONCustom(msg)

	switch output {
	case "text":
		marshaller = msg.MarshalTextIndent

	case "json":
		marshaller = marshalJSONCustom(msg)
	}

	data, err := marshaller()
	if err != nil {
		log.Errorf("Error to marshal message: %s", err)
	}

	title := titleStd(fmt.Sprintf("%s (%s output)", name, output))
	log.Infof("%s\n%s", title, string(data))
}

func PrintStruct(log Logger, title string, i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	log.Infof("%s\n%s", titleStd(title), s)
}

func titleStd(t string) string {
	return title(t, 45)
}

// title returns title with left and right fillers.
func title(title string, length int) string {
	const filler = "-"

	fillLen := length - len(title) - 2
	if fillLen < 2 {
		return title
	}

	fillRightLen := fillLen / 2
	fillLeftLen := fillRightLen + (fillLen % 2)

	return fmt.Sprintf(
		"%s %s %s",
		strings.Repeat(filler, fillLeftLen),
		title,
		strings.Repeat(filler, fillRightLen),
	)
}
