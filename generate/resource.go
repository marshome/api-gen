package generate

import (
	"fmt"

	"github.com/marshome/apis/spec"
)

type Resource struct {
	c         *Context
	name      string
	parent    string
	m         *spec.APIResource
	resources []*Resource

	Methods []*Method
}

func (r *Resource) CacheRequestTypes() {
	for _, meth := range r.Methods {
		meth.CacheRequestTypes()
	}
	for _, res := range r.resources {
		res.CacheRequestTypes()
	}
}

func (r *Resource) CacheResponseTypes() {
	for _, meth := range r.Methods {
		meth.CacheResponseTypes()
	}

	if r.resources != nil {
		for _, res := range r.resources {
			res.CacheResponseTypes()
		}
	}
}

func (r *Resource) GoField() string {
	return r.c.InitialCap(r.name)
}

func (r *Resource) GoName() string {
	return r.c.InitialCap(fmt.Sprintf("%s.%s", r.parent, r.name))
}

func (r *Resource) GoType() string {
	return r.GoName() + "Service"
}

func (r *Resource) ParseMethods() {
	r.Methods = []*Method{}
	if r.m.Methods == nil {
		return
	}

	methMap := make(map[string]interface{})
	for k, v := range r.m.Methods {
		methMap[k] = v
	}

	for _, mname := range r.c.SortedKeys(methMap) {
		mi := r.m.Methods[mname]
		r.Methods = append(r.Methods, &Method{
			c:    r.c,
			r:    r,
			name: mname,
			doc:  mi,
		})
	}

	for _, v := range r.resources {
		v.ParseMethods()
	}
}