package codegen

import (
	"github.com/marshome/apis/spec"
	"encoding/json"
	"bytes"
	"go/format"
)

func GenerateClient(specJson []byte, namespace string) (code string, err error) {
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

	c.GenerateClientImports()

	c.Pn("const(")
	c.Pn("    RootUrl= \"%s\"", c.spec.RootUrl)
	c.Pn(")")

	c.GenerateSchemas()

	//service new()
	c.Pn("func New(client *http.Client) (*Service_, error) {")
	c.Pn("    if client == nil {")
	c.Pn("        return nil, errors.New(\"client is nil\")")
	c.Pn("    }")
	c.Pn("")
	c.Pn("s := &Service_{client: client, RootUrl:RootUrl }")
	c.Pn("")
	for _, r := range c.Resources {
		c.Pn("    s.%s = New%sService(s)", GoName(r.Name, true), GoName(r.Name, true))
	}
	c.Pn("")
	c.Pn("    return s,nil")
	c.Pn("}")
	c.Pn("")

	//service def
	c.Pn("type Service_ struct {")
	c.Pn("    client *http.Client")
	c.Pn("    RootUrl string")
	c.Pn("")
	for _, r := range c.Resources {
		c.Pn("%s *%sService", GoName(r.Name, true), GoName(r.Name, true))
	}
	c.Pn("}")
	c.Pn("")

	for _, r := range c.Resources {
		r.GenerateClient()
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}