package generate

import (
	"github.com/marshome/apis/googlespec"
)

type Method struct {
	c                     *Context
	r                     *Resource
	name                  string
	spec                   *googlespec.APIMethod

	GoName                string
	OptionParamStructType string
	RequestType           string
	ResponseType          string
	PathParams            []*Param
	RequiredQueryParams   []*Param
	OptionalQueryParams     []*Param
	Signature string
}

func NewMethod(ctx *Context,r *Resource,name string,spec *googlespec.APIMethod)(m *Method) {
	m = &Method{
		c:ctx,
		r:r,
		name:name,
		spec:spec,
	}

	m.GoName = Depunct(m.name, true)

	m.PathParams = make([]*Param, 0)
	m.RequiredQueryParams = make([]*Param, 0)
	if m.spec.ParameterOrder != nil {
		for _, name := range m.spec.ParameterOrder {
			p := NewParam(ctx, m, name, m.spec.Parameters[name])
			if p.m.Location == "path" {
				m.PathParams = append(m.PathParams, p)
			} else if p.m.Location == "query" && p.m.Required {
				m.RequiredQueryParams = append(m.RequiredQueryParams, p)
			}
		}
	}

	m.OptionQueryParams = make([]*Param, 0)
	optionParamMap := make(map[string]interface{})
	for k, v := range m.spec.Parameters {
		if v.Required {
			continue
		}

		optionParamMap[k] = v
	}
	for _, name := range m.c.SortedKeys(optionParamMap) {
		mi := m.spec.Parameters[name]
		m.OptionQueryParams = append(m.OptionQueryParams, NewParam(ctx, m, name, mi))
	}

	if m.r == nil {
		m.OptionParamStructType = "Service_" + m.GoName + "Options"
	} else {
		m.OptionParamStructType = m.r.GoName + "_" + m.GoName + "Options"
	}

	m.RequestType = m.buildRequestType()
	m.ResponseType = m.buildResponseType()
	m.Signature = m.buildSignature()

	return m
}

func (m*Method)buildRequestType()string {
	if m.spec.Request == nil {
		return ""
	} else {
		if s := m.c.Schemas[m.spec.Request.Ref]; s != nil {
			return "*" + s.GoName
		} else {
			panic("ref not found " + m.spec.Request.Ref)
		}
	}
}

func (m* Method)buildResponseType()string {
	if m.spec.Response == nil {
		return ""
	} else {
		if s := m.c.Schemas[m.spec.Response.Ref]; s != nil {
			return "*" + s.GoName
		} else {
			panic("ref not found " + m.spec.Response.Ref)
		}
	}
}

func (m *Method) buildSignature()(sig string) {
	//context
	sig += "    " + m.GoName + "(_ctx *marsapi.Context"

	//path params
	for _, param := range m.PathParams {
		sig += "," + param.GoName + " " + param.GoType
	}

	//required query params
	for _, param := range m.RequiredQueryParams {
		sig += "," + param.GoName + " " + param.GoType
	}

	//body param
	reqType := m.RequestType
	if reqType != "" {
		sig += ", req" + reqType
	}

	//option params
	if len(m.OptionalQueryParams) > 0 {
		sig += ",_opt *" + m.OptionParamStructType
	}

	//response
	sig += ")"
	resType := m.ResponseType
	if resType == "" {
		sig += "(err error)"
	} else {
		sig += "(resp " + resType + ", err error)"
	}

	return sig
}