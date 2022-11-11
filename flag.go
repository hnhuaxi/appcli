package appcli

import (
	"errors"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type _Flag struct {
	Name        string
	Type        string
	Category    string
	DefaultText string
	FilePath    string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool
	Aliases    []string
	EnvVars    []string

	Base int
}

type Flag struct {
	cli.Flag
	As string // name 可以改名
}

func (flag *Flag) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return ErrInvalidMapNode
	}
	if len(value.Content) < 2 {
		return ErrMapMemberNotEnough
	}

	var (
		_flag _Flag
	)

	if err := value.Decode(&_flag); err != nil {
		return err
	}

	var (
		typ   = lookupNodeString(value.Content, "type")
		cflag cli.Flag
		err   error
	)

	cflag, err = unmarshalFlag(strings.ToLower(typ), value)
	if err != nil {
		return err
	}

	flag.Flag = cflag

	return nil
}

func lookupNodeString(content []*yaml.Node, field string) string {
	for i := 0; i < len(content); i += 2 {
		var key, val = content[i], content[i+1]
		if key.Value == field {
			return val.Value
		}
	}

	return ""
}

func lookupNodeValue(content []*yaml.Node, field string) *yaml.Node {
	for i := 0; i < len(content); i += 2 {
		var key, val = content[i], content[i+1]
		if key.Value == field {
			return val
		}
	}

	return nil
}

func unmarshalFlag(typ string, value *yaml.Node) (cli.Flag, error) {
	switch typ {
	case "bool":
		return decode[*cli.BoolFlag](value)
	case "duration":
		return decode[*cli.DurationFlag](value)
	case "float64":
		return decode[*cli.Float64Flag](value)
	case "float64slice":
		return decode[*cli.Float64SliceFlag](value)
	case "int":
		return decode[*cli.IntFlag](value)
	case "int64":
		return decode[*cli.Int64Flag](value)
	case "intslice":
		return decode[*cli.IntSliceFlag](value)
	case "path":
		return decode[*cli.PathFlag](value)
	case "string":
		return decode[*cli.StringFlag](value)
	case "stringslice":
		return decode[*cli.StringSliceFlag](value)
	case "timestamp":
		return decode[*cli.TimestampFlag](value)
	case "uint":
		return decode[*cli.UintFlag](value)
	case "uint64":
		return decode[*cli.Uint64Flag](value)
	default:
		return nil, errors.New("nonimplement type")
	}
}

func decode[T cli.Flag](value *yaml.Node) (flag T, err error) {
	var (
		z T
		t = reflect.TypeOf(z)
		v = reflect.New(t.Elem())
	)

	if err = value.Decode(v.Interface()); err != nil {
		return z, err
	}

	return v.Interface().(T), nil
}
