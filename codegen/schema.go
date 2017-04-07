package codegen

import (
	"github.com/marshome/apis/spec"
	"fmt"
)

type Schema struct {
	c    *Context
	Spec *spec.Object
	Name string
	Top bool
}

func NewSchema(c *Context, spec *spec.Object,name string,top bool)*Schema {
	s := &Schema{
		c:c,
		Spec:spec,
		Name:name,
		Top:top,
	}

	return s
}

func (s *Schema)GenerateCode() {
	if s.Spec.Desc != "" {
		s.c.Comment(fmt.Sprintf("%s :%s", GoName(s.Name, true), s.Spec.Desc))
	}

	if !s.Top {
		s.c.Pn("type %s struct{", GoName(s.Name, true))
		for _, p := range s.Spec.Fields {
			if p.Desc != "" {
				s.c.Comment(fmt.Sprintf("%s :%s", GoName(p.Name, true), p.Desc))
			}
			//s.c.AddFieldValueComments(s.c.P, p, "\t", p.spec.Description != "")

			s.c.Pn(" %s %s `json:\"%s,omitempty\"`", GoName(p.Name, true), GoType(p, s.Name, false), GoName(p.Name, false))
			s.c.Pn("")
		}
		s.c.Pn("}")
		s.c.Pn("")
	} else {
		if s.Spec.Collection == spec.COLLECTION_NONE {
			s.c.Pn("type %s %s", GoName(s.Name, true), GoType(s.Spec, "", false))
		} else {
			s.c.Pn("type %s %s", GoName(s.Name, true), GoType(s.Spec, "", true))
		}
	}
}