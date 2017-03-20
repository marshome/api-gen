package generate

import (
	"log"
	"strings"

	"github.com/marshome/apis/spec"
)

type Method struct {
	c         *Context
	r         *Resource
	name      string

	doc       *spec.APIMethod

	params    []*Param

	arguments *Arguments
}

func (m *Method) Id() string {
	return m.doc.Id
}

func (m *Method)GoName()string{
	return m.c.InitialCap(m.name)
}

func (m *Method) supportsMediaUpload() bool {
	return m.doc.MediaUpload != nil
}

func (m *Method) mediaUploadPath() string {
	if m.doc.MediaUpload == nil {
		return ""
	}

	if m.doc.MediaUpload.Protocols == nil {
		return ""
	}

	if m.doc.MediaUpload.Protocols.Simple == nil {
		return ""
	}

	return m.doc.MediaUpload.Protocols.Simple.Path
}

func (m *Method) supportsMediaDownload() bool {
	if m.supportsMediaUpload() {
		// storage.objects.insert claims support for download in
		// addition to upload but attempting to do so fails.
		// This situation doesn't apply to any other methods.
		return false
	}
	return m.doc.SupportsMediaDownload
}

func (m *Method) Params() []*Param {
	if m.params == nil {
		paramMap := make(map[string]interface{})
		for k, v := range m.doc.Parameters {
			paramMap[k] = v
		}
		for _, name := range m.c.SortedKeys(paramMap) {
			mi := m.doc.Parameters[name]
			m.params = append(m.params, &Param{
				name:   name,
				m:      mi,
				method: m,
			})
		}
	}
	return m.params
}

func (m *Method) grepParams(f func(*Param) bool) []*Param {
	matches := make([]*Param, 0)
	for _, param := range m.Params() {
		if f(param) {
			matches = append(matches, param)
		}
	}
	return matches
}

func (m *Method) NamedParam(name string) *Param {
	matches := m.grepParams(func(p *Param) bool {
		return p.name == name
	})
	if len(matches) < 1 {
		log.Panicf("failed to find named parameter %q", name)
	}
	if len(matches) > 1 {
		log.Panicf("found multiple parameters for parameter name %q", name)
	}
	return matches[0]
}

func (m *Method) OptParams() []*Param {
	return m.grepParams(func(p *Param) bool {
		return !p.IsRequired()
	})
}

func (m *Method)OptionsType()string {
	if m.r == nil {
		return "Service_" + m.GoName() + "Options"
	} else {
		return m.r.GoName()+"_" + m.GoName() + "Options"
	}
}

func (meth *Method) CacheRequestTypes() {
	if reqType := meth.GetRequestType(); reqType != "" && strings.HasPrefix(reqType, "*") {
		meth.c.RequestTypes[reqType] = true
	}
}

func (meth *Method) CacheResponseTypes() {
	if retType := meth.GetResponseType(); retType != "" && strings.HasPrefix(retType, "*") {
		meth.c.ResponseTypes[retType] = true
	}
}

func (meth *Method) GetRequestType() (typ string) {
	if meth.doc.Request == nil {
		return ""
	} else {
		if s := meth.c.Schemas[meth.doc.Request.Ref]; s != nil {
			return s.GoReturnType()
		} else {
			return "*" + meth.doc.Request.Ref
		}
	}
}

func (meth *Method) GetResponseType() (typ string) {
	if meth.doc.Response == nil {
		return ""
	} else {
		if s := meth.c.Schemas[meth.doc.Response.Ref]; s != nil {
			return s.GoReturnType()
		} else {
			return "*" + meth.doc.Response.Ref
		}
	}
}

func (meth *Method) Signature()(sig string) {
	//context
	sig += "    " + meth.GoName() + "(_ctx *marsapi.Context"

	//path params
	for _, param := range meth.NewArguments().l {
		if param.location == "path" {
			sig += "," + param.goname + " " + param.gotype
		}
	}

	//required query params
	for _, param := range meth.NewArguments().l {
		if param.location == "query" &&param.required {
			sig += "," + param.goname + " " + param.gotype
		}
	}

	//body
	reqType := meth.GetRequestType()
	if reqType != "" {
		sig += ", req" + reqType
	}

	//options
	if len(meth.OptParams()) > 0 {
		sig += ",_opt " + meth.OptionsType()
	}

	//response
	sig += ")"
	resType := meth.GetResponseType()
	if resType == "" {
		sig += "(err error)"
	} else {
		sig += "(resp " + resType + ", err error)"
	}

	return sig
}

func (meth *Method) NewArguments() (args *Arguments) {
	if meth.arguments == nil {
		args = &Arguments{
			method: meth,
			m:      make(map[string]*Argument),
		}

		if meth.doc.ParameterOrder != nil {
			for _, pname := range meth.doc.ParameterOrder {
				arg := meth.NewArg(pname, meth.NamedParam(pname))
				args.AddArg(arg)
			}
		}

		if meth.doc.Request != nil {
			args.AddArg(meth.NewBodyArg(meth.doc.Request.Ref))
		}

		meth.arguments = args
	}

	return meth.arguments
}

func (meth *Method) NewBodyArg(ref string) *Argument {
	return &Argument{
		goname:   meth.c.ValidGoIdentifer(strings.ToLower(ref)),
		apiname:  "REQUEST",
		gotype:   "*" + ref,
		apitype:  ref,
		location: "body",
		desc:     "",
	}
}

func (meth *Method) NewArg(apiname string, p *Param) *Argument {
	m := p.m
	apitype := m.Type
	des := m.Description
	goname := meth.c.ValidGoIdentifer(apiname) // but might be changed later, if conflicts
	if strings.Contains(des, "identifier") && !strings.HasSuffix(strings.ToLower(goname), "id") {
		goname += "id" // yay
		p.callFieldName = goname
	}
	gotype := meth.c.MustSimpleTypeConvert(apitype, m.Format)
	if p.IsRepeated() {
		gotype = "[]" + gotype
	}
	return &Argument{
		apiname:  apiname,
		apitype:  apitype,
		goname:   goname,
		gotype:   gotype,
		location: m.Location,
		desc:     des,
		required:p.IsRequired(),
	}
}

// Strips the leading '*' from a type name so that it can be used to create a literal.
func (meth *Method) ResponseTypeLiteral() string {
	v := meth.GetResponseType()
	if strings.HasPrefix(v, "*") {
		return v[1:]
	}
	return v
}

func (meth *Method)PathParams()(paramList []*Argument) {
	paramList = make([]*Argument, 0)

	args := meth.NewArguments()
	for _, arg := range args.m {
		if arg.location == "path" {
			paramList = append(paramList, arg)
		}
	}

	return paramList
}