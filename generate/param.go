package generate

import (
	"fmt"

	"github.com/marshome/apis/googlespec"
)

type Param struct {
	c             *Context
	method        *Method
	name          string
	m             *googlespec.APIObject

	GoName string
	GoType string
}

func NewParam(c *Context,m *Method,name string,spec *googlespec.APIObject) *Param {
	p := &Param{
		c:c,
		method:m,
		name:name,
		m:spec,
	}

	p.GoName=p.buildGoName()
	p.GoType = p.buildGoType()

	return p
}

func (p* Param)buildGoName()string {
	s := Depunct(p.name, true)
	for _, v := range go_tokens {
		if s == v {
			s = v + "_"
			break
		}
	}

	return s
}

func (p *Param) buildGoType() string {
	t, ok := p.c.SimpleTypeConvert(p.m.Type, p.m.Format)
	if !ok {
		panic("failed to convert parameter type " + fmt.Sprintf("type=%q, format=%q", p.m.Type, p.m.Format))
	}

	if p.m.Location == "query" &&p.m.Repeated {
		t = "[]" + t
	}

	return t
}
