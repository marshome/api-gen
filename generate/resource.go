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

func (r *Resource) GenerateType() {
	pn := r.c.Pn
	t := r.GoType()
	pn(fmt.Sprintf("func New%s(s *Service) *%s {", t, t))
	pn("rs := &%s{s : s}", t)
	for _, res := range r.resources {
		pn("rs.%s = New%s(s)", res.GoField(), res.GoType())
	}
	pn("return rs")
	pn("}")

	pn("\ntype %s struct {", t)
	pn(" s *Service")
	for _, res := range r.resources {
		pn("\n\t%s\t*%s", res.GoField(), res.GoType())
	}
	pn("}")

	for _, res := range r.resources {
		res.GenerateType()
	}
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

func (r *Resource) GenerateClientMethods() {
	for _, meth := range r.Methods {
		meth.GenerateClientCode()
	}
	for _, res := range r.resources {
		res.GenerateClientMethods()
	}
}

func (r *Resource) GoField() string {
	return r.c.InitialCap(r.name)
}

func (r *Resource) GoType() string {
	return r.c.InitialCap(fmt.Sprintf("%s.%s", r.parent, r.name)) + "Service"
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

func (r *Resource) GenerateServerMethods() {
	r.c.GenerateServerService(r, r.Methods)

	for _, res := range r.resources {
		res.GenerateServerMethods()
	}
}
