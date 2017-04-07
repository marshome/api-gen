package codegen

import (
	"github.com/marshome/apis/spec"
)

type Resource struct {
	c  *Context
	Spec *spec.Resource
	Name string

	Methods []*Method
	SubResources []*Resource
}

func NewResource(c *Context,spec *spec.Resource,name string)*Resource {
	r := &Resource{
		c:c,
		Spec:spec,
		Name:name,
	}

	if spec.Methods != nil {
		r.Methods = make([]*Method, 0)
		for _, m := range spec.Methods {
			r.Methods = append(r.Methods, NewMethod(c,r, m))
		}
	}

	if spec.Resources != nil {
		r.SubResources = make([]*Resource, 0)
		for _, subR := range spec.Resources {
			r.SubResources = append(r.SubResources, NewResource(c, subR, name + "." + subR.Name))
		}
	}

	return r
}

func (r *Resource)GenerateService() {
	serviceName := GoName(r.Name, true) + "Service"

	//def
	if r.Spec.Desc != "" {
		r.c.Comment(r.Spec.Desc)
	}
	r.c.Pn("type %s interface{", serviceName)
	for _, m := range r.Methods {
		m.GenerateComments()
		r.c.Pn(m.GenerateSignature())
	}
	r.c.Pn("}")
	r.c.Pn("")

	//default impl
	r.c.Pn("type Default%s_ struct{", serviceName)
	r.c.Pn("}")
	r.c.Pn("")
	for _, m := range r.Methods {
		r.c.Pn("func (s *Default%s_) %s{", serviceName, m.GenerateSignature())
		if m.Spec.Response==""{
			r.c.Pn("    return nil")
		}else{
			r.c.Pn("    return nil,nil")
		}
		r.c.Pn("}")
		r.c.Pn("")
	}

	//options
	for _, m := range r.Methods {
		m.GenerateOptionalParams()
	}

	//router
	r.c.Pn("func Handle%s(_r marsapi.Router,_service %s)(err error){", serviceName, serviceName)
	for _, m := range r.Methods {
		m.GenerateRouter()
	}
	r.c.Pn("    return nil")
	r.c.Pn("}")

	if r.SubResources != nil {
		for _, subResource := range r.SubResources {
			subResource.GenerateService()
		}
	}
}

func (r *Resource)GenerateClient() {
	//new()
	r.c.Pn("func New%sService(s *Service_) *%sService {", GoName(r.Name, true), GoName(r.Name, true))
	r.c.Pn("    rs:=&%sService{s:s}", GoName(r.Name, true))
	r.c.Pn("")
	if r.SubResources != nil {
		for _, sub := range r.SubResources {
			r.c.Pn("    rs.%s=New%sService(s)", GoName(sub.Name, true), GoName(sub.Name, true))
		}
	}
	r.c.Pn("")
	r.c.Pn("    return rs")
	r.c.Pn("}")
	r.c.Pn("")

	//def
	r.c.Pn("type %sService struct{", GoName(r.Name, true))
	r.c.Pn("    s *Service_")
	r.c.Pn("")
	if r.SubResources != nil {
		for _, sub := range r.SubResources {
			r.c.Pn("    %s *%sService", GoName(sub.Name, true), GoName(sub.Name, true))
		}
	}
	r.c.Pn("}")
	r.c.Pn("")

	//method call
	if r.Methods != nil {
		for _, m := range r.Methods {
			m.GenerateClientCall()
		}
	}

	if r.SubResources != nil {
		for _, subResource := range r.SubResources {
			subResource.GenerateClient()
		}
	}
}