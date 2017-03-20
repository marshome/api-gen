package generate

import (
	"fmt"
	"log"
	"os"

	"github.com/marshome/apis/spec"
)

type Schema struct {
	c *Context

	m *spec.APIObject

	typ *Type // lazily populated by Type

	apiName      string // the native API-defined name of this type
	goName       string // lazily populated by GoName
	goReturnType string // lazily populated by GoReturnType

	properties []*Property
}

func NewSchema(ctx *Context,spec *spec.APIObject,apiName string, typ  *Type) (s*Schema) {
	s = &Schema{
		c:ctx,
		m:spec,
		apiName:apiName,
		typ:typ,
	}

	if s.typ == nil {
		s.typ = &Type{c: s.c, m: s.m, _apiName: s.apiName}
	}

	if name, ok := s.Type().MapType(); ok {
		s.goName = name
	} else {
		s.goName = s.c.GetName(s.c.InitialCap(s.apiName))
	}

	if s.Type().IsMap() {
		s.goReturnType = s.GoName()
	} else {
		s.goReturnType = "*" + s.GoName()
	}

	if s.Type().IsStruct() {
		pl := []*Property{}
		propMap := make(map[string]interface{})
		for k, v := range s.m.Properties {
			propMap[k] = v
		}
		for _, name := range s.c.SortedKeys(propMap) {
			m := s.m.Properties[name]
			pl = append(pl, &Property{
				s:       s,
				m:       m,
				apiName: name,
			})
		}

		s.properties = pl
	}

	return s
}

func (s *Schema) Type() *Type {
	return s.typ
}

func (s *Schema) GoName() string {
	return s.goName
}

func (s *Schema) GoReturnType() string {
	return s.goReturnType
}

func (s *Schema) GetProperties() []*Property {
	return s.properties
}

func (s *Schema) populateSubSchemas() (outerr error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		outerr = fmt.Errorf("%v", r)
	}()

	addSubStruct := func(subApiName string, t *Type) {
		if s.c.Schemas[subApiName] != nil {
			panic("dup schema apiName: " + subApiName)
		}
		subm := t.m
		subs:=NewSchema(s.c,subm,subApiName,t)
		s.c.Schemas[subApiName] = subs
		err := subs.populateSubSchemas()
		if err != nil {
			s.c.Panicf("in sub-struct %q: %v", subApiName, err)
		}
	}

	if s.Type().IsStruct() {
		for _, p := range s.GetProperties() {
			if p.Type().IsSimple() || p.Type().IsMap() {
				continue
			}
			if at, ok := p.Type().ArrayType(); ok {
				if at.IsSimple() || at.IsReference() {
					continue
				}
				subApiName := fmt.Sprintf("%s.%s", s.apiName, p.apiName)
				if at.IsStruct() {
					addSubStruct(subApiName, at) // was p.Type()?
					continue
				}
				if _, ok := at.ArrayType(); ok {
					addSubStruct(subApiName, at)
					continue
				}
				s.c.Panicf("Unknown property array type for %q: %s", subApiName, at)
				continue
			}
			subApiName := fmt.Sprintf("%s.%s", s.apiName, p.apiName)
			if p.Type().IsStruct() {
				addSubStruct(subApiName, p.Type())
				continue
			}
			if p.Type().IsReference() {
				continue
			}
			s.c.Panicf("Unknown type for %q: %s", subApiName, p.Type())
		}
		return
	}

	if at, ok := s.Type().ArrayType(); ok {
		if at.IsSimple() || at.IsReference() {
			return
		}
		subApiName := fmt.Sprintf("%s.Item", s.apiName)

		if at.IsStruct() {
			addSubStruct(subApiName, at)
			return
		}
		if at, ok := at.ArrayType(); ok {
			if at.IsSimple() || at.IsReference() {
				return
			}
			addSubStruct(subApiName, at)
			return
		}
		s.c.Panicf("Unknown array type for %q: %s", subApiName, at)
		return
	}

	if s.Type().IsSimple() || s.Type().IsReference() {
		return
	}

	fmt.Fprintf(os.Stderr, "in populateSubSchemas, schema is: %s", s.c.PrettyJSON(s.m))
	s.c.Panicf("populateSubSchemas: unsupported type for schema %q", s.apiName)
	panic("unreachable")
}

func (s *Schema) WriteSchemaCode() {
	if s.Type().IsAny() {
		s.c.Pn("\ntype %s interface{}", s.GoName())
		return
	}
	if s.Type().IsStruct() && !s.Type().IsMap() {
		s.writeSchemaStruct()
		return
	}

	if _, ok := s.Type().ArrayType(); ok {
		log.Printf("TODO writeSchemaCode for arrays for %s", s.GoName())
		return
	}

	if destSchema, ok := s.Type().ReferenceSchema(); ok {
		// Convert it to a struct using embedding.
		s.c.Pn("\ntype %s struct {", s.GoName())
		s.c.Pn(" %s", destSchema.GoName())
		s.c.Pn("}")
		return
	}

	if s.Type().IsSimple() {
		apitype := s.m.Type
		typ := s.c.MustSimpleTypeConvert(apitype, s.m.Format)
		s.c.Pn("\ntype %s %s", s.GoName(), typ)
		return
	}

	if s.Type().IsMap() {
		return
	}

	fmt.Fprintf(os.Stderr, "in writeSchemaCode, schema is: %s", s.c.PrettyJSON(s.m))
	s.c.Panicf("writeSchemaCode: unsupported type for schema %q", s.apiName)
}

func (s *Schema) Description() string {
	return s.m.Description
}

func (s *Schema) writeSchemaStruct() {
	s.c.P("\n")
	des := s.Description()
	if des != "" {
		s.c.P("%s", s.c.AsComment("", fmt.Sprintf("%s: %s", s.GoName(), des)))
	}
	s.c.Pn("type %s struct {", s.GoName())

	np := new(namePool)
	if s.isResponseType() {
		np.Get("ServerResponse") // reserve the name
	}

	firstFieldName := "" // used to store a struct field name for use in documentation.
	for i, p := range s.GetProperties() {
		if i > 0 {
			s.c.P("\n")
		}

		pname := np.Get(p.GoName())
		if pname[0] == '@' {
			continue
		}

		des := p.Description()
		if des != "" {
			s.c.P("%s", s.c.AsComment("\t", fmt.Sprintf("%s: %s", pname, des)))
		}
		s.c.AddFieldValueComments(s.c.P, p, "\t", des != "")

		var extraOpt string
		if p.Type().IsIntAsString() {
			extraOpt += ",string"
		}

		typ := p.Type().AsGo()

		s.c.Pn(" %s %s `json:\"%s,omitempty%s\"`", pname, typ, p.APIName(), extraOpt)
		if firstFieldName == "" {
			firstFieldName = pname
		}
	}

	if s.isResponseType() {
		if firstFieldName != "" {
			s.c.P("\n")
		}
		s.c.P("%s", s.c.AsComment("\t", "ServerResponse contains the HTTP response code and headers from the server."))
		s.c.P(" marsapi.ServerResponse `json:\"-\"`")
	}

	s.c.Pn("}")
	s.c.Pn("")

	return
}

func (s *Schema) isResponseType() bool {
	return s.c.ResponseTypes["*"+s.goName]
}