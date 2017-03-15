package generate

import (
	"fmt"
	"strings"

	"github.com/marshome/apis/spec"
)

type Param struct {
	method        *Method
	name          string
	m             *spec.APIObject
	callFieldName string // empty means to use the default
}

func (p *Param) Default() string {
	return p.m.Default
}

func (p *Param) Enum() []string {
	return p.m.Enum
}

func (p *Param) EnumDescriptions() []string {
	return p.m.EnumDescriptions
}

func (p *Param) UnfortunateDefault() bool {
	return false
}

func (p *Param) IsRequired() bool {
	return p.m.Required
}

func (p *Param) IsRepeated() bool {
	return p.m.Repeated
}

func (p *Param) Location() string {
	return p.m.Location
}

func (p *Param) GoType() string {
	typ, format := p.m.Type, p.m.Format
	if typ == "string" && strings.Contains(format, "int") && p.Location() != "query" {
		panic("unexpected int parameter encoded as string, not in query: " + p.name)
	}
	t, ok := p.method.c.SimpleTypeConvert(typ, format)
	if !ok {
		panic("failed to convert parameter type " + fmt.Sprintf("type=%q, format=%q", typ, format))
	}
	return t
}
