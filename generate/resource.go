package generate

import (
	"fmt"

	"github.com/marshome/apis/googlespec"
)

type Resource struct {
	c         *Context
	name      string
	parent    string
	m         *googlespec.APIResource
	resources []*Resource

	GoName    string
	GoType string

	Methods   []*Method
}

func NewResource(c *Context,name string,parentName string,spec *googlespec.APIResource,subResources []*Resource)*Resource {
	r := &Resource{
		c:c,
		name:name,
		parent:parentName,
		m:spec,
		resources:subResources,
	}

	r.GoName = r.c.InitialCap(fmt.Sprintf("%s.%s", r.parent, r.name))
	r.GoType = r.GoName + "Service"

	r.parseMethods()

	return r
}

func (r *Resource) parseMethods() {
	r.Methods = []*Method{}
	if r.m.Methods == nil {
		return
	}

	methodMap := make(map[string]interface{})
	for k, v := range r.m.Methods {
		methodMap[k] = v
	}

	for _, name := range r.c.SortedKeys(methodMap) {
		spec := r.m.Methods[name]
		r.Methods = append(r.Methods, NewMethod(
			r.c,
			r,
			name,
			spec,
		))
	}

	for _, v := range r.resources {
		v.parseMethods()
	}
}