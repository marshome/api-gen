package generate

import (
	"log"
	"strconv"

	"github.com/marshome/apis/spec"
	"fmt"
)

type Property struct {
	s       *Schema             // property of which schema
	apiName string              // the native API-defined name of this property
	m       *spec.APIObject // original JSON map

	typ *Type // lazily populated by Type
}

func (p *Property) Type() *Type {
	if p.typ == nil {
		p.typ = &Type{c: p.s.c, m: p.m, _apiName:fmt.Sprintf("%s.%s", p.s.apiName, p.apiName)}
	}
	return p.typ
}

func (p *Property) GoName() string {
	return p.s.c.InitialCap(p.apiName)
}

func (p *Property) APIName() string {
	return p.apiName
}

func (p *Property) Default() string {
	return p.m.Default
}

func (p *Property) Description() string {
	return p.m.Description
}

func (p *Property) Enum() []string {
	if p.m.Enum != nil {
		return p.m.Enum
	}

	// Check if this has an array of string enums.
	if items := p.m.Items; items != nil {
		if items.Enum != nil && items.Type == "string" {
			return items.Enum
		}
	}

	return nil
}

func (p *Property) EnumDescriptions() []string {
	if p.m.EnumDescriptions != nil {
		return p.m.EnumDescriptions
	}

	// Check if this has an array of string enum descriptions.
	if items := p.m.Items; items != nil {
		if desc := p.m.Items.EnumDescriptions; desc != nil {
			return desc
		}
	}
	return nil
}

func (p *Property) Pattern() string {
	return p.m.Pattern
}

// UnfortunateDefault reports whether p may be set to a zero value, but has a non-zero default.
func (p *Property) UnfortunateDefault() bool {
	switch p.Type().AsGo() {
	default:
		return false

	case "bool":
		return p.Default() == "true"

	case "string":
		if p.Default() == "" {
			return false
		}
		// String fields are considered to "allow" a zero value if either:
		//  (a) they are an enum, and one of the permitted enum values is the empty string, or
		//  (b) they have a validation pattern which matches the empty string.
		pattern := p.Pattern()
		enum := p.Enum()
		if pattern != "" && enum != nil {
			log.Printf("Encountered enum property which also has a pattern: %#v", p)
			return false // don't know how to handle this, so ignore.
		}
		return (pattern != "" && p.s.c.EmptyPattern(pattern)) ||
			(enum != nil && p.s.c.EmptyEnum(enum))

	case "float64", "int64", "uint64", "int32", "uint32":
		if p.Default() == "" {
			return false
		}
		if f, err := strconv.ParseFloat(p.Default(), 64); err == nil {
			return f != 0.0
		}
		// The default value has an unexpected form.  Whatever it is, it's non-zero.
		return true
	}
}