package codegen

import (
	"bytes"
	"encoding/json"
	"go/format"
	"github.com/marshome/i-api/spec"
)

func GenerateServer(specJson []byte) (code []byte, err error) {
	doc := spec.Document{}
	err = json.Unmarshal(specJson, &doc)
	if err != nil {
		return nil, err
	}

	c := Context{}
	c.Code = &bytes.Buffer{}
	c.spec = doc

	c.Parse()

	c.Pn("package %s", Namespace(c.spec.Name, c.spec.Version))
	c.Pn("")

	c.GenerateServerImports()

	c.GenerateSchemas()

	for _, r := range c.Resources {
		r.GenerateService()
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.Bytes(), err
	}
	return clean, nil
}