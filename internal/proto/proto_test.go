package proto

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFile = "testdata/example.proto"

var testfiles = []string{testFile}

func TestProto_NewProto_Success(t *testing.T) {
	p, err := NewProto(testfiles)
	assert.NotNil(t, p)
	assert.Nil(t, err)
}

func TestProto_NewProto_HTTP_Request(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(testFile)
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(w, file)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	p, err := NewProto([]string{ts.URL})
	assert.NotNil(t, p)
	assert.Nil(t, err)
}

func TestProto_NewProto_NotFound(t *testing.T) {
	var files = []string{"testdata/not_found.proto"}

	p, err := NewProto(files)
	assert.Nil(t, p)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf("open %s: no such file or directory", files[0]), err.Error())
	}
}

func TestProto_NewProto_ParseError(t *testing.T) {
	var files = []string{"testdata/parse_err.proto"}

	p, err := NewProto(files)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestProto_FindMessage(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"found", args{"HelloRequest"}, "HelloRequest", false},
		{"found", args{"example.HelloRequest"}, "HelloRequest", false},
		{"not found", args{"NotFound"}, "", true},
	}

	p, err := NewProto(testfiles)
	assert.Nil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.FindMessage(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Proto.FindMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil && got.GetName() != tt.want {
				t.Errorf("Proto.FindMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
