package proto

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

type Proto struct {
	descriptors []*desc.FileDescriptor
	ImportPaths []string
}

// NewProto creates a new instance of Proto. May return a Proto or open/parse error.
func NewProto(filenames []string, importPaths ...string) (p *Proto, err error) {
	importPaths = append(importPaths, ".")
	importPaths = append(importPaths, build.Default.SrcDirs()...)
	importPaths = append(importPaths, "/")

	parser := protoparse.Parser{
		ImportPaths: importPaths,
	}

	// resolve filenames: local filename save as is, remote files are downloads and saves as temp files
	paths := make([]string, 0, len(filenames))
	for _, f := range filenames {
		u, err := url.Parse(f)
		if err != nil {
			return nil, err
		}

		if u.Scheme == "http" || u.Scheme == "https" { // url
			filename, cleaner, err := httpGet(f)
			if err != nil {
				return nil, err
			}
			defer cleaner()

			paths = append(paths, filename)
		} else if p, _ := filepath.Glob(f); len(p) > 0 { // check pattern
			paths = append(paths, p...)
		} else { // path
			paths = append(paths, f)
		}
	}

	descriptors, err := parser.ParseFiles(paths...)
	if err != nil {
		return
	}

	return &Proto{
		descriptors: descriptors,
		ImportPaths: importPaths,
	}, nil
}

// FindMessage searches for message with given name.
func (p *Proto) FindMessage(name string) (*desc.MessageDescriptor, error) {
	if name == "" {
		return nil, errors.New("proto: name is empty")
	}

	for _, fd := range p.descriptors {
		// finds the message with the given fully-qualified name
		if d := fd.FindMessage(name); d != nil {
			return d, nil
		}

		// find just with short-name
		for _, d := range fd.GetMessageTypes() {
			if d.GetName() == name {
				return d, nil
			}
		}
	}

	return nil, fmt.Errorf("proto: message with name %s not found", name)
}

func httpGet(url string) (filename string, cleaner func(), err error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp("", "protokaf.*.proto")
	if err != nil {
		return
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return
	}

	return f.Name(), func() { os.Remove(f.Name()) }, f.Sync()
}
