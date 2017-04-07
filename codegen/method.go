package codegen

import (
	"github.com/marshome/apis/spec"
	"fmt"
	"strings"
)

type Method struct {
	c      *Context
	Resource *Resource
	Spec     *spec.Method
}

func NewMethod(c *Context,r *Resource,spec *spec.Method)*Method {
	m := &Method{
		c:c,
		Resource:r,
		Spec:spec,
	}

	return m
}

func (m *Method)OptionalParamName()string {
	return GoName(m.Resource.Name, true) + "_" + GoName(m.Spec.Name, true) + "Params"
}

func (m *Method)GenerateComments() {
	if m.Spec.Desc != "" {
		m.c.Comment(m.Spec.Desc)
	}

	if m.Spec.PathParams != nil&&len(m.Spec.PathParams) > 0 {
		m.c.Comment("-PATH PARAMS:")
		for _, p := range m.Spec.PathParams {
			m.c.Comment(fmt.Sprintf("   @%s :%s", GoName(p.Name, false), p.Desc))
		}
	}

	if m.Spec.RequiredQueryParams != nil&&len(m.Spec.RequiredQueryParams) > 0 {
		m.c.Comment("-REQUIRED QUERY PARAMS:")
		for _, p := range m.Spec.RequiredQueryParams {
			m.c.Comment(fmt.Sprintf("   @%s :%s", GoName(p.Name, false), p.Desc))
		}
	}
}

func (m *Method)GenerateSignature()string {
	//name
	sig := GoName(m.Spec.Name, true)

	//ctx
	sig += "(_ctx *marsapi.Context"

	//path params
	for _, p := range m.Spec.PathParams {
		sig += "," + GoName(p.Name, false) + " " + GoType(p, "", false)
	}

	//required query params
	for _, p := range m.Spec.RequiredQueryParams {
		sig += "," + GoName(p.Name, false) + " " + GoType(p, "", false)
	}

	//body param
	if m.Spec.Request != "" {
		sig += ", req *" + GoName(m.Spec.Request, true)
	}

	//option params
	if len(m.Spec.OptionalQueryParams) > 0 {
		sig += ",_opt *" + m.OptionalParamName()
	}

	sig += ")"

	//response
	if m.Spec.Response == "" {
		sig += "(err error)"
	} else {
		sig += "(resp *" + GoName(m.Spec.Response, true) + ", err error)"
	}

	return sig
}

func (m *Method)GenerateOptionalParams() {
	if len(m.Spec.OptionalQueryParams) == 0 {
		return
	}

	//def
	m.c.Pn("type %s struct{", m.OptionalParamName())
	for _, p := range m.Spec.OptionalQueryParams {
		if p.Desc != "" {
			m.c.Comment(fmt.Sprintf("%s :%s", GoName(p.Name, true), p.Desc))
		}
		m.c.Pn("%s *%s", GoName(p.Name, true), GoType(p, "", false))
	}
	m.c.Pn("}")
	m.c.Pn("")

	//parse func
	m.c.Pn("func Parse%s(values url.Values)(_opts *%s,_err error){", m.OptionalParamName(), m.OptionalParamName())
	m.c.Pn("    _opts=&%s{}", m.OptionalParamName())
	m.c.Pn("")
	m.c.Pn("    var _s string")
	m.c.Pn("")
	for _, p := range m.Spec.OptionalQueryParams {
		onError := func() {
			m.c.Pn("if _err!=nil{")
			m.c.Pn("    return nil,_err")
			m.c.Pn("}")
			m.c.Pn("")
		}

		goName := GoName(p.Name, true)
		m.c.Pn("/* param:%s */", p.Name)
		m.c.Pn("_s=values.Get(\"%s\")", p.Name)
		m.c.Pn("if _s!=\"\"{")
		if p.Collection == spec.COLLECTION_NONE {
			if p.Type == spec.TYPE_STRING {
				m.c.Pn("_opts.%s=marsapi.String(_s)", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_BOOL {
				m.c.Pn("_p,_err:=strconv.ParseBool(_s)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Bool(_p)", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_BYTE {
				m.c.Pn("_p_int64,_err:=strconv.ParseInt(_s,10,64)")
				onError()
				m.c.Pn("if _p_int64<0||_p_int64>255{", )
				m.c.Pn("    _ctx.ServiceError=errors.New(\"byte out of range:%s\")", p.Name)
				m.c.Pn("    return")
				m.c.Pn("}")
				m.c.Pn("_opts.%s=marsapi.Byte(byte(_p_int64))", goName)
			} else if p.Type == spec.TYPE_INT32 {
				m.c.Pn("_p,_err:=strconv.ParseInt(_s,10,32)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Int32(int32(_p))", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_UINT32 {
				m.c.Pn("_p,_err:=strconv.ParseUint(_s,10,32)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Uint32(uint32(_p))", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_INT64 {
				m.c.Pn("_p,_err:=strconv.ParseInt(_s,10,64)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Int64(_p)", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_UINT64 {
				m.c.Pn("_p,_err:=strconv.ParseUint(_s,10,64)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Uint64(_p)", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_FLOAT32 {
				m.c.Pn("_p,_err:=strconv.ParseFloat(_s,32)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Float32(float32(_p))", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_FLOAT64 {
				m.c.Pn("_p,_err:=strconv.ParseFloat(_s,64)")
				onError()
				m.c.Pn("_opts.%s=marsapi.Float64(_p)", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_DATE {
				m.c.Pn("_p,_err:=time.Parse(time.RFC3339,_s)")
				onError()
				m.c.Pn("_opts.%s=&_p", goName)
				m.c.Pn("")
			} else if p.Type == spec.TYPE_DATETIME {
				m.c.Pn("_p,_err:=time.Parse(time.RFC3339,_s)")
				onError()
				m.c.Pn("_opts.%s=&_p", goName)
				m.c.Pn("")
			} else {
				panic("unkown option query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
					",param=" + p.Name + " " + p.Type)
			}
		} else if p.Collection == spec.COLLECTION_ARRAY {
			if p.Type == spec.TYPE_STRING {
				m.c.Pn("_p,_err:=marsapi.ParseStringList(_s)")
				onError()
			} else if p.Type == spec.TYPE_INT32 {
				panic("1")
			} else if p.Type == spec.TYPE_UINT32 {
				panic("2")
			} else if p.Type == spec.TYPE_INT64 {
				panic("3")
			} else if p.Type == spec.TYPE_UINT64 {
				panic("4")
			} else if p.Type == spec.TYPE_FLOAT32 {
				panic("5")
			} else if p.Type == spec.TYPE_FLOAT64 {
				panic("6")
			} else {
				panic("unkown option query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
					",param=" + p.Name + " " + p.Type)
			}
		} else {
			panic("unkown option query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
				",param=" + p.Name + " " + p.Collection)
		}
		m.c.Pn("    }")
		m.c.Pn("")
	}
	m.c.Pn("    return _opts,nil")
	m.c.Pn("}")
	m.c.Pn("")
}

func (m *Method)GenerateRouter() {
	onError := func() {
		m.c.Pn("if _err!=nil{")
		m.c.Pn("    _ctx.ServiceError=_err")
		m.c.Pn("    return")
		m.c.Pn("}")
	}

	//handle func
	m.c.Pn("    _r.Handle(\"%s\",\"%s\", func(_ctx *marsapi.Context) {", m.Spec.HttpMethod, m.Spec.Path)
	m.c.Pn("    var _err error")
	m.c.Pn("")
	if len(m.Spec.PathParams) > 0 || len(m.Spec.RequiredQueryParams) > 0 {
		m.c.Pn("var _s string")
	}
	m.c.Pn("")

	//path params
	if m.Spec.PathParams != nil&&len(m.Spec.PathParams) > 0 {
		m.c.Pn("/****** path params ******/")
		for _, p := range m.Spec.PathParams {
			m.c.Pn("/* path param:%s */", p.Name)
			m.c.Pn("_s=_ctx.PathParamMap[\"%s\"]", p.Name)
			m.c.Pn("if _s==\"\"{")
			m.c.Pn("    _ctx.ServiceError=errors.New(\"Missing:%s\")", p.Name)
			m.c.Pn("    return")
			m.c.Pn("}")

			goName := GoName(p.Name, false)

			if p.Collection == spec.COLLECTION_NONE {
				if p.Type == spec.TYPE_STRING {
					m.c.Pn("_p_%s:=_s", goName, )
				} else if p.Type == spec.TYPE_BYTE {
					m.c.Pn("_p_%s_int64,_err:=strconv.ParseInt(_s,10,64)", goName)
					onError()
					m.c.Pn("if _p_%s_int64<0||_p_%s_int64>255{", goName, goName)
					m.c.Pn("    _ctx.ServiceError=errors.New(\"byte out of range:%s\")", goName)
					m.c.Pn("    return")
					m.c.Pn("}")
					m.c.Pn("_p_%s:=byte(_p_%s_int64)", goName, goName)
				} else if p.Type == spec.TYPE_INT32 {
					m.c.Pn("_p_%s_int64,_err:=strconv.ParseInt(_s,10,32)", goName)
					onError()
					m.c.Pn("_p_%s:=int32(_p_%s_int64)", goName, goName)
				} else if p.Type == spec.TYPE_UINT32 {
					m.c.Pn("_p_%s_uint64,_err:=strconv.ParseUint(_s,10,32)", goName)
					onError()
					m.c.Pn("_p_%s:=int32(_p_%s_uint64)", goName, goName)
				} else if p.Type == spec.TYPE_INT64 {
					m.c.Pn("_p_%s,_err:=strconv.ParseInt(_s,10,64)", goName)
					onError()
				} else if p.Type == spec.TYPE_UINT64 {
					m.c.Pn("_p_%s,_err:=strconv.ParseUint(_s,10,64)", goName)
					onError()

				} else {
					panic("unknown path param type,meth=" + m.Spec.Name + ",param=" + p.Name + ",type=" + p.Type)
				}
			} else {
				panic("collection type in path param not support " + m.Spec.Name + " " + p.Name + " " + p.Collection)
			}

			m.c.Pn("")
		}
	}

	//required query params
	if m.Spec.RequiredQueryParams != nil&&len(m.Spec.RequiredQueryParams) > 0 {
		m.c.Pn("/****** required query params ******/")
		for _, p := range m.Spec.RequiredQueryParams {
			m.c.Pn("/* required query param:%s */", p.Name)
			m.c.Pn("_s=_ctx.HttpRequest.URL.Query().Get(\"%s\")", p.Name)
			m.c.Pn("if _s==\"\"{")
			m.c.Pn("    _ctx.ServiceError=errors.New(\"Missing:%s\")", p.Name)
			m.c.Pn("    return")
			m.c.Pn("}")

			goName := GoName(p.Name, false)

			if p.Collection == spec.COLLECTION_NONE {
				if p.Type == spec.TYPE_STRING {
					m.c.Pn("_q_%s:=_s", goName)
				} else if p.Type == spec.TYPE_BOOL {
					m.c.Pn("_q_%s,_err:=strconv.ParseBool(_s)", goName)
					onError()
				} else if p.Type == spec.TYPE_BYTE {
					m.c.Pn("_q_%s_int64,_err:=strconv.ParseInt(_s,10,64)", goName)
					onError()
					m.c.Pn("if _q_%s_int64<0||_q_%s_int64>255{", goName, goName)
					m.c.Pn("    _ctx.ServiceError=errors.New(\"byte out of range:%s\")", goName)
					m.c.Pn("    return")
					m.c.Pn("}")
					m.c.Pn("_q_%s:=byte(_q_%s_int64)", goName, goName)
				} else if p.Type == spec.TYPE_INT32 {
					m.c.Pn("_q_%s_int64,_err:=strconv.ParseInt(_s,10,32)", goName)
					onError()
					m.c.Pn("_q_%s:=int32(_q_%s_int64)", goName, goName)
				} else if p.Type == spec.TYPE_UINT32 {
					m.c.Pn("_q_%s_uint64,_err:=strconv.ParseUint(_s,10,32)", goName)
					onError()
					m.c.Pn("_q_%s:=uint32(_q_%s_uint64)", goName, goName)
				} else if p.Type == spec.TYPE_INT64 {
					m.c.Pn("_q_%s,_err:=strconv.ParseInt(_s,10,64)", goName)
					onError()
				} else if p.Type == spec.TYPE_UINT64 {
					m.c.Pn("_q_%s,_err:=strconv.ParseUint(_s,10,64)", goName)
					onError()
				} else if p.Type == spec.TYPE_FLOAT32 {
					m.c.Pn("_q_%s,_err:=strconv.ParseFloat(_s,32)", goName)
					onError()
				} else if p.Type == spec.TYPE_FLOAT64 {
					m.c.Pn("_q_%s,_err:=strconv.ParseFloat(_s,64)", goName)
					onError()
				} else if p.Type == spec.TYPE_DATE {
					m.c.Pn("_q_%s,_err:=time.Parse(time.RFC3339,_s)", goName)
					onError()
				} else if p.Type == spec.TYPE_DATETIME {
					m.c.Pn("_q_%s,_err:=time.Parse(time.RFC3339,_s)", goName)
					onError()
				} else {
					panic("unkown required query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
						",param=" + p.Name + " " + p.Type)
				}
			} else if p.Collection == spec.COLLECTION_ARRAY {
				if p.Type == spec.TYPE_STRING {
					m.c.Pn("_q_%s,_err:=marsapi.ParseStringList(_s)", goName)
					onError()
				} else if p.Type == spec.TYPE_INT32 {
					panic("1")
				} else if p.Type == spec.TYPE_UINT32 {
					panic("2")
				} else if p.Type == spec.TYPE_INT64 {
					panic("3")
				} else if p.Type == spec.TYPE_UINT64 {
					panic("4")
				} else if p.Type == spec.TYPE_FLOAT32 {
					panic("5")
				} else if p.Type == spec.TYPE_FLOAT64 {
					panic("6")
				} else {
					panic("unkown required query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
						",param=" + p.Name + " " + p.Type)
				}
			} else {
				panic("unkown required query param type,meth=" + m.Resource.Name + "_" + m.Spec.Name +
					",param=" + p.Name + " " + p.Collection)
			}

			m.c.Pn("")
		}
	}

	//request body
	if m.Spec.Request != "" {
		m.c.Pn("/****** request body ******/")
		m.c.Pn("    body,_err:= ioutil.ReadAll(_ctx.HttpRequest.Body)")
		onError()
		m.c.Pn("    _req:=%s{}", m.Spec.Request)
		m.c.Pn("    _err=json.Unmarshal(body,&_req)")
		onError()
		m.c.Pn("")
	}

	//optional query params
	if len(m.Spec.OptionalQueryParams) > 0 {
		m.c.Pn("/****** optional query params ******/")
		m.c.Pn("    _opts,_err:=Parse%s(_ctx.HttpRequest.URL.Query())", m.OptionalParamName())
		onError()
		m.c.Pn("")
	}

	//call handler
	m.c.Pn("/****** call handler ******/")
	call := ""
	if m.Spec.Response != "" {
		call += "        _resp,_err:="
	} else {
		call += "        _err="
	}
	call += fmt.Sprintf("_service.%s(_ctx", GoName(m.Spec.Name, true))
	for _, p := range m.Spec.PathParams {
		call += fmt.Sprintf(",_p_%s", GoName(p.Name, false))
	}
	for _, p := range m.Spec.RequiredQueryParams {
		call += fmt.Sprintf(",_q_%s", GoName(p.Name, false))
	}
	if m.Spec.Request != "" {
		call += ",&_req"
	}
	if len(m.Spec.OptionalQueryParams) > 0 {
		call += ",_opts"
	}
	call += ")"
	m.c.Pn(call)
	m.c.Pn("")
	m.c.Pn("/****** result ******/")
	if m.Spec.Response != "" {
		m.c.Pn("    _ctx.ServiceResponse=_resp")
	}
	m.c.Pn("    _ctx.ServiceError=_err")
	//method options
	m.c.Pn("    },&marsapi.MethodOptions{})")
	m.c.Pn("")
}

func (m *Method)GenerateClientCall() {
	callName := fmt.Sprintf("%s_%sCall", GoName(m.Resource.Name, true), GoName(m.Spec.Name, true))

	//def
	m.c.Pn("type %s struct{", callName)
	m.c.Pn("    marsapi.ApiCall")
	m.c.Pn("")
	m.c.Pn("    s *Service_")
	m.c.Pn("}")
	m.c.Pn("")

	//new()
	sig := ""
	if m.Spec.PathParams != nil {
		for _, p := range m.Spec.PathParams {
			sig += fmt.Sprintf("%s %s,", GoName(p.Name, false), GoType(p, "", false))
		}
	}
	if m.Spec.RequiredQueryParams != nil {
		for _, p := range m.Spec.RequiredQueryParams {
			sig += fmt.Sprintf("%s %s,", GoName(p.Name, false), GoType(p, "", false))
		}
	}
	if m.Spec.Request != "" {
		sig += fmt.Sprintf("req_ *%s", GoName(m.Spec.Request, true))
	}
	m.c.Pn("func (r *%sService)%s(%s)*%s{", GoName(m.Resource.Name, true), GoName(m.Spec.Name, true), sig, callName)
	m.c.Pn("c:=&%s{",callName)
	m.c.Pn("    s:r.s,")
	m.c.Pn("    ApiCall:*marsapi.NewApiCall(r.s.client,RootUrl,\"%s\"),", m.Spec.Path)
	m.c.Pn("    }")
	m.c.Pn("")
	m.c.Pn("    c.HttpMethod_=\"%s\"",strings.ToUpper(m.Spec.HttpMethod))
	if m.Spec.PathParams != nil&&len(m.Spec.PathParams) > 0 {
		m.c.Pn("c.PathParams_=make(map[string]string)")
		for _, p := range m.Spec.PathParams {
			if GoType(p,"",false)=="string"{
				m.c.Pn("c.PathParams_[\"%s\"]=%s", GoName(p.Name, false), GoName(p.Name, false))
			}else{
				m.c.Pn("c.PathParams_[\"%s\"]=fmt.Sprint(%s)", GoName(p.Name, false), GoName(p.Name, false))
			}
		}
	}
	if m.Spec.RequiredQueryParams != nil {
		for _, p := range m.Spec.RequiredQueryParams {
			if GoType(p,"",false)=="string" {
				m.c.Pn("c.QueryParams_.Set(\"%s\",%s)", GoName(p.Name, false), GoName(p.Name, false))
			}else{
				m.c.Pn("c.QueryParams_.Set(\"%s\",fmt.Sprint(%s))", GoName(p.Name, false), GoName(p.Name, false))
			}
		}
	}
	if m.Spec.Request != "" {
		m.c.Pn("    c.BodyParams_=req_")
	}
	m.c.Pn("")
	m.c.Pn("    return c")
	m.c.Pn("}")
	m.c.Pn("")

	//optional query params
	if m.Spec.OptionalQueryParams != nil {
		for _, p := range m.Spec.OptionalQueryParams {
			typ := GoType(p, "", false)
			m.c.Pn("func (c *%s)Set%s(%s %s)*%s{", callName, GoName(p.Name, true), GoName(p.Name, false), typ, callName)
			if typ == "string" {
				m.c.Pn("    c.QueryParams_.Set(\"%s\",%s)", GoName(p.Name, false), GoName(p.Name, false))
			} else {
				m.c.Pn("    c.QueryParams_.Set(\"%s\",fmt.Sprint(%s))", GoName(p.Name, false), GoName(p.Name, false))
			}
			m.c.Pn("")
			m.c.Pn("    return c")
			m.c.Pn("}")
			m.c.Pn("")
		}
	}

	//do
	if m.Spec.Response == "" {
		m.c.Pn("func (c *%s)Do()(err error){", callName)
	} else {
		m.c.Pn("func (c *%s)Do()(resp *%s,err error){", callName, GoName(m.Spec.Response, true))
	}
	m.c.Pn("    err=c.SendRequest_()")
	m.c.Pn("    if err!=nil{")
	if m.Spec.Response == "" {
		m.c.Pn("    return err")
	} else {
		m.c.Pn("    return nil,err")
	}
	m.c.Pn("    }")
	m.c.Pn("")
	if m.Spec.Response == "" {
		m.c.Pn("return nil")
	} else {
		m.c.Pn("return c.ResponseData_.(*%s),nil", GoName(m.Spec.Response, true))
	}
	m.c.Pn("}")
	m.c.Pn("")
}