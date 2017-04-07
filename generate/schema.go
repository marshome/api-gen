package generate

import (
	"fmt"

	"github.com/marshome/apis/googlespec"
)

type Schema struct {
	c          *Context
	m          *googlespec.APIObject
	name    string

	GoName     string
	Type       *Type
	Properties []*Property
}

func NewSchema(ctx *Context,spec *googlespec.APIObject,name string, typ  *Type) (s*Schema) {
	s = &Schema{
		c:ctx,
		m:spec,
		name:name,
		Type:typ,
	}

	if s.Type == nil {
		s.Type = &Type{c: s.c, m: s.m, _apiName: s.name}
	}

	s.GoName = s.c.InitialCap(s.name)

	if s.Type.IsStruct() {
		pl := []*Property{}
		propMap := make(map[string]interface{})
		for k, v := range s.m.Properties {
			propMap[k] = v
		}
		for _, name := range s.c.SortedKeys(propMap) {
			m := s.m.Properties[name]
			pl = append(pl, NewProperty(s, name, m))
		}

		s.Properties = pl
	}

	return s
}

func parseSchemaFromRight(ctx*Context, t* Type, name string)(subSchema *Schema) {
	if t.IsSimple() || t.IsReference(){
		return nil
	}

	if t.IsStruct() {
		return NewSchema(ctx, t.m, name, t)
	} else if at := t.ArrayType(); at != nil {
		return parseSchemaFromRight(ctx, at, name)
	} else if mt := t.MapType(); mt != nil {
		return parseSchemaFromRight(ctx, mt, name)
	} else {
		panic("unknown array type " + t._apiName)
	}
}

func (s *Schema) ParseSubSchemas(schemas map[string]*Schema) {
	add := func(subs *Schema) {
		if oldSchema := schemas[subs.name]; oldSchema != nil {
			if oldSchema.Type.IsStruct()&&oldSchema.Type.ArrayType() == nil&&oldSchema.Type.MapType() == nil {
				panic("sub schema already exist: " + subs.name)
			}
		}

		schemas[subs.name] = subs
	}

	if mt := s.Type.MapType(); mt != nil {
		subs := parseSchemaFromRight(s.c, mt, s.name)
		if subs != nil {
			add(subs)
			subs.ParseSubSchemas(schemas)
		}
	} else if at := s.Type.ArrayType(); at != nil {
		subs := parseSchemaFromRight(s.c, at, s.name)
		if subs != nil {
			add(subs)
			subs.ParseSubSchemas(schemas)
		}
	} else if s.Type.IsStruct() {
		for _, p := range s.Properties {
			if p.Type.IsSimple() || p.Type.IsReference() {
				continue
			}

			name := fmt.Sprintf("%s.%s", s.name, p.name)
			if p.Type.IsStruct() {
				subs := NewSchema(s.c, p.Type.m, name, p.Type)
				add(subs)
				subs.ParseSubSchemas(schemas)
			} else if at := p.Type.ArrayType(); at != nil {
				subs := parseSchemaFromRight(s.c, at, name)
				if subs == nil {
					continue
				}
				add(subs)
				subs.ParseSubSchemas(schemas)
			} else if mt := p.Type.MapType(); mt != nil {
				subs := parseSchemaFromRight(s.c, mt, name)
				if subs == nil {
					continue
				}
				add(subs)
				subs.ParseSubSchemas(schemas)
			} else {
				panic(fmt.Sprintf("Unknown type for %s : %s", name, p.Type))
			}
		}
	}//nothing
}

func (s *Schema) GenerateSchema() {
	panic("111")
	if s.Type.IsSimple() {
		typ, ok := s.c.SimpleTypeConvert(s.m.Type, s.m.Format)
		if !ok {
			panic(fmt.Sprintf("SimpleTypeConvert failed type=%s,format=%s", s.m.Type, s.m.Format))
		}
		s.c.Pn("type %s %s", s.GoName, typ)
		s.c.Pn("")
	} else if mt := s.Type.MapType(); mt != nil {
		s.c.Pn("type %s map[string]%s",s.name,mt.GoType())
		s.c.Pn("")
	} else if s.Type.IsStruct() {
		if s.m.Description != "" {
			s.c.P("%s", s.c.AsComment("", fmt.Sprintf("%s: %s", s.GoName, s.m.Description)))
		}
		s.c.Pn("type %s struct {", s.GoName)
		for _, p := range s.Properties {
			if p.GoName[0] == '@' {
				continue
			}

			if p.spec.Description != "" {
				s.c.P("%s", s.c.AsComment("\t", fmt.Sprintf("%s: %s", p.GoName, p.spec.Description)))
			}
			s.c.AddFieldValueComments(s.c.P, p, "\t", p.spec.Description != "")

			var extraOpt string
			if p.Type.IsIntAsString() {
				extraOpt += ",string"
			}
			s.c.Pn(" %s %s `json:\"%s,omitempty%s\"`", p.GoName, p.Type.GoType(), p.name, extraOpt)
			s.c.Pn("")
		}
		s.c.Pn("}")
		s.c.Pn("")
	} else {
		panic(fmt.Sprintf("GenerateSchema: unsupported type for schema %q", s.name))
	}
}
