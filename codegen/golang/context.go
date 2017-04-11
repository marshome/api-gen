package codegen

import (
	"bytes"
	"github.com/marshome/i-api/spec"
	"fmt"
	"sort"
)

var go_tokens = map[string]int{
	"break":1,
	"case":1,
	"chan":1,
	"const":1,
	"continue":1,
	"default":1,
	"defer":1,
	"else":1,
	"fallthrough":1,
	"for":1,
	"func":1,
	"go":1,
	"goto":1,
	"if":1,
	"import":1,
	"interface":1,
	"map":1,
	"package":1,
	"range":1,
	"return":1,
	"select":1,
	"struct":1,
	"switch":1,
	"type":1,
	"var":1,
}

type Context struct {
	spec      spec.Document
	Code      *bytes.Buffer

	Schemas   map[string]*Schema
	Resources []*Resource
}

func (c *Context) Pn(format string, args ...interface{}) {
	c.Code.WriteString(fmt.Sprintf(format + "\n", args...))
}

func (c *Context)Comment(s string) {
	c.Pn("%s", Comment("", fmt.Sprintf(" %s", s)))
}

func (c *Context)GenerateServerImports() {
	c.Pn("import \"io/ioutil\"")
	c.Pn("import \"encoding/json\"")
	c.Pn("import \"errors\"")
	c.Pn("import \"net/http\"")
	c.Pn("import \"net/url\"")
	c.Pn("import \"strconv\"")
	c.Pn("import \"time\"")
	c.Pn("import \"github.com/marshome/i-api/genlib\"")
	c.Pn("")

	c.Pn("var _=errors.New(\"\")")
	c.Pn("var _=http.DefaultClient")
	c.Pn("var _=&url.URL{}")
	c.Pn("var _=strconv.ErrRange")
	c.Pn("var _=ioutil.Discard")
	c.Pn("var _=json.InvalidUTF8Error{}")
	c.Pn("var _=genlib.Bool(false)")
	c.Pn("var _=time.RFC3339")
	c.Pn("")
}

func (c *Context)GenerateClientImports() {
	c.Pn("import \"io/ioutil\"")
	c.Pn("import \"encoding/json\"")
	c.Pn("import \"errors\"")
	c.Pn("import \"bytes\"")
	c.Pn("import \"net/http\"")
	c.Pn("import \"net/url\"")
	c.Pn("import \"strconv\"")
	c.Pn("import \"time\"")
	c.Pn("import \"fmt\"")
	c.Pn("import \"context\"")
	c.Pn("import \"io\"")
	c.Pn("import \"github.com/marshome/i-api/genlib\"")
	c.Pn("")

	c.Pn("var _=errors.New(\"\")")
	c.Pn("var _=bytes.ErrTooLarge")
	c.Pn("var _=http.DefaultClient")
	c.Pn("var _=&url.URL{}")
	c.Pn("var _=strconv.ErrRange")
	c.Pn("var _=ioutil.Discard")
	c.Pn("var _=json.InvalidUTF8Error{}")
	c.Pn("var _=genlib.Bool(false)")
	c.Pn("var _=time.RFC3339")
	c.Pn("var _=context.Canceled")
	c.Pn("var _=io.EOF")
	c.Pn("var _,_=fmt.Print(\"\")")
	c.Pn("")

}

func (c *Context)parseAllSchemas(s *spec.Object, name string, top bool) {
	if top && s.Collection != spec.COLLECTION_NONE {
		//top collection
		c.Schemas[name] = NewSchema(c, s, name, true)
		if s.Type == spec.TYPE_OBJECT {
			c.Schemas[name + "Item"] = NewSchema(c, s, name + "Item", false)
			if s.Fields != nil {
				for _, f := range s.Fields {
					if f.Type == spec.TYPE_OBJECT {
						c.Schemas[name + "Item." + f.Name] = NewSchema(c, f, name + "Item." + f.Name, false)
					}//nothing
					c.parseAllSchemas(f, name + "Item." + f.Name, false)
				}
			}//nothing
		}//nothing

		if s.CollectionItem != nil {
			c.parseAllSchemas(s.CollectionItem, name + "Item." + s.CollectionItem.Name, false)
		}//nothing

		return
	}//nothing

	if s.CollectionItem != nil {
		c.parseAllSchemas(s.CollectionItem, name + "." + s.CollectionItem.Name, false)
	}//nothing

	if s.Type == spec.TYPE_OBJECT {
		c.Schemas[name] = NewSchema(c, s, name, false)
		if s.Fields != nil {
			for _, f := range s.Fields {
				if f.Type == spec.TYPE_OBJECT {
					c.Schemas[name + "." + f.Name] = NewSchema(c, f, name + "." + f.Name, false)
				}//nothing
				c.parseAllSchemas(f, name + "." + f.Name, false)
			}
		}//nothing
	} else {
		if top {
			c.Schemas[name] = NewSchema(c, s, name, true)
		}//nothing
	}
}

func (c *Context)Parse() {
	if c.spec.Schemas != nil {
		c.Schemas = make(map[string]*Schema)
		for _, v := range c.spec.Schemas {
			c.parseAllSchemas(v, v.Name, true)
		}
	}

	if c.spec.Resources != nil {
		c.Resources = make([]*Resource, 0)
		for _, v := range c.spec.Resources {
			c.Resources = append(c.Resources, NewResource(c, v, v.Name))
		}
	}
}

func (c *Context)GenerateSchemas() {
	if c.Schemas == nil {
		return
	}

	schemaList := make([]*Schema, 0)
	for _, v := range c.Schemas {
		schemaList = append(schemaList, v)
	}
	sort.SliceStable(schemaList, func(i, j int) bool {
		return schemaList[i].Name < schemaList[j].Name
	})

	for _, v := range schemaList {
		v.GenerateCode()
	}
}