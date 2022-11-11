package appcli

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	OutputFormat string
)

const (
	OutJSON  OutputFormat = "json"
	OutPlain OutputFormat = "plain"
)

type Output struct {
	Format OutputFormat
}

type Encoder interface {
	Encode(val interface{}) error
}

type plainEncoder struct {
	w io.Writer
}

func NewPlainEncoder(w io.Writer) *plainEncoder {
	return &plainEncoder{w}
}

func (plain *plainEncoder) Encode(val interface{}) error {
	_, err := io.WriteString(plain.w, fmt.Sprint(val))
	return err
}

func (outfmt OutputFormat) NewEncode(w io.Writer) Encoder {
	switch outfmt {
	case "json":
		return json.NewEncoder(w)
	case "plain":
		return NewPlainEncoder(w)
	default:
		return nil
	}
}
