package codegen

import (
	"bytes"
	"encoding/json"
	"go/format"
	"github.com/marshome/apis/spec"
)

func GenerateServer(specJson []byte, namespace string) (code string, err error) {
	doc := spec.Document{}
	err = json.Unmarshal(specJson, &doc)
	if err != nil {
		return "", err
	}

	c := Context{}
	c.Namespace = namespace
	c.Code = &bytes.Buffer{}
	c.spec = doc

	c.Parse()

	c.Pn("package %s", c.Namespace)
	c.Pn("")

	c.GenerateServerImports()

	c.GenerateSchemas()

	for _, r := range c.Resources {
		r.GenerateService()
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}