package generate

import (
	"fmt"
	"github.com/marshome/apis/googlespec"
)

type Property struct {
	s    *Schema
	name string
	spec    *googlespec.APIObject

	GoName string
	Type  *Type
	Enum  []string
	EnumDescriptions []string
}

func NewProperty(s *Schema,name string,spec    *googlespec.APIObject)*Property {
	p := &Property{
		s:s,
		name:name,
		spec:spec,
	}

	p.GoName = p.s.c.InitialCap(p.name)
	p.Type = &Type{c: p.s.c, m: p.spec, _apiName:fmt.Sprintf("%s.%s", p.s.name, p.name)}
	p.Enum=p.buildEnum()
	p.EnumDescriptions=p.buildEnumDescriptions()

	return p
}

func (p *Property) buildEnum() []string {
	if p.spec.Enum != nil {
		return p.spec.Enum
	}

	// Check if this has an array of string enums.
	if items := p.spec.Items; items != nil {
		if items.Enum != nil && items.Type == "string" {
			return items.Enum
		}
	}

	return nil
}

func (p *Property) buildEnumDescriptions() []string {
	if p.spec.EnumDescriptions != nil {
		return p.spec.EnumDescriptions
	}

	// Check if this has an array of string enum descriptions.
	if items := p.spec.Items; items != nil {
		if desc := p.spec.Items.EnumDescriptions; desc != nil {
			return desc
		}
	}
	return nil
}