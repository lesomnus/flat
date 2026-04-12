package configs

import (
	"io"
	"reflect"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/lesomnus/z"
)

type configmap map[string]any

func (m configmap) unmarshal(kind string, node ast.Node) (any, error) {
	t, ok := m[kind]
	if !ok {
		return nil, io.EOF
	}

	v := reflect.New(reflect.TypeOf(t)).Interface()
	if node == nil {
		return v, nil
	}

	raw, err := node.MarshalYAML()
	if err != nil {
		return nil, z.Err(err, "marshal to raw")
	}
	if err := yaml.Unmarshal(raw, v); err != nil {
		return nil, z.Err(err, "decode: %q", kind)
	}

	return v, nil
}
