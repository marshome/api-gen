package generate

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/marshome/apis/googlespec"
)

var (
	BaseURL = flag.String("base_url", "", "(optional) Override the default service API URL. If empty, the service's root URL will be used.")
)

var go_tokens=[]string{
	"break",
	"case",
	"chan",
	"const",
	"continue",
	"default",
	"defer",
	"else",
	"fallthrough",
	"for",
	"func",
	"go",
	"goto",
	"if",
	"import",
	"interface",
	"map",
	"package",
	"range",
	"return",
	"select",
	"struct",
	"switch",
	"type",
	"var",
}

type ServerGenerateParams struct {
	Namespace string
}

type Context struct {
	ApiPackageBase string
	Namespace string

	Doc            *googlespec.APIDocument
	Code           *bytes.Buffer

	usedNames     namePool

	Schemas    map[string]*Schema
	Resources  []*Resource
}

// namePool keeps track of used names and assigns free ones based on a
// preferred name
type namePool struct {
	m map[string]bool // lazily initialized
}

func (p *namePool) Get(preferred string) string {
	if p.m == nil {
		p.m = make(map[string]bool)
	}
	name := preferred
	tries := 0
	for p.m[name] {
		tries++
		name = fmt.Sprintf("%s%d", preferred, tries)
	}
	p.m[name] = true
	return name
}

func (c *Context) P(format string, args ...interface{}) {
	c.Code.WriteString(fmt.Sprintf(format, args...))
}

func (c *Context) Pn(format string, args ...interface{}) {
	c.Code.WriteString(fmt.Sprintf(format + "\n", args...))
}

func (c *Context) SortedKeys(m map[string]interface{}) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}

func Depunct(ident string, needCap bool) string {
	var buf bytes.Buffer
	preserve_ := false
	for i, c := range ident {
		if c == '_' {
			if preserve_ || strings.HasPrefix(ident[i:], "__") {
				preserve_ = true
			} else {
				needCap = true
				continue
			}
		} else {
			preserve_ = false
		}
		if c == '-' || c == '.' || c == '$' || c == '/' {
			needCap = true
			continue
		}
		if needCap {
			c = unicode.ToUpper(c)
			needCap = false
		}
		buf.WriteByte(byte(c))
	}
	return buf.String()

}

// initialCap returns the identifier with a leading capital letter.
// it also maps "foo-bar" to "FooBar".
func (c *Context) InitialCap(ident string) string {
	if ident == "" {
		panic("blank identifier")
	}
	return Depunct(ident, true)
}

var urlRE = regexp.MustCompile(`^http\S+$`)

func (ctx *Context) AsComment(pfx, c string) string {
	var buf bytes.Buffer
	const maxLen = 70
	r := strings.NewReplacer(
		"\n", "\n"+pfx+"// ",
		"`\"", `"`,
		"\"`", `"`,
	)
	for len(c) > 0 {
		line := c
		if len(line) < maxLen {
			fmt.Fprintf(&buf, "%s// %s\n", pfx, r.Replace(line))
			break
		}
		// Don't break URLs.
		if !urlRE.MatchString(line[:maxLen]) {
			line = line[:maxLen]
		}
		si := strings.LastIndex(line, " ")
		if nl := strings.Index(line, "\n"); nl != -1 && nl < si {
			si = nl
		}
		if si != -1 {
			line = line[:si]
		}
		fmt.Fprintf(&buf, "%s// %s\n", pfx, r.Replace(line))
		c = c[len(line):]
		if si != -1 {
			c = c[1:]
		}
	}
	return buf.String()
}

// GetName returns a free top-level function/type identifier in the package.
// It tries to return your preferred match if it's free.
func (c *Context) GetName(preferred string) string {
	return c.usedNames.Get(preferred)
}

func (c *Context) AddFieldValueComments(p func(format string, args ...interface{}), field *Property, indent string, blankLine bool) {
	var lines []string

	if field.Enum != nil {
		desc := field.EnumDescriptions
		lines = append(lines, c.AsComment(indent, "Possible values:"))
		defval := field.spec.Default
		for i, v := range field.Enum {
			more := ""
			if v == defval {
				more = " (default)"
			}
			if len(desc) > i && desc[i] != "" {
				more = more + " - " + desc[i]
			}
			lines = append(lines, c.AsComment(indent, `  "`+v+`"`+more))
		}
	}
	if blankLine && len(lines) > 0 {
		p(indent + "//\n")
	}
	for _, l := range lines {
		p("%s", l)
	}
}

func (c *Context) SimpleTypeConvert(apiType, format string) (gotype string, ok bool) {
	// From http://tools.ietf.org/html/draft-zyp-json-schema-03#section-5.1
	switch apiType {
	case "boolean":
		gotype = "bool"
	case "string":
		gotype = "string"
		switch format {
		case "int64", "uint64", "int32", "uint32":
			gotype = format
		}
	case "number":
		gotype = "float64"
	case "integer":
		gotype = "int64"
	case "any":
		gotype = "interface{}"
	}
	return gotype, gotype != ""
}

func (c *Context) PrettyJSON(m interface{}) string {
	bs, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("[JSON error %v on %#v]", err, m)
	}
	return string(bs)
}

func (c *Context) EmptyPattern(pattern string) bool {
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString("")
	}
	log.Printf("Encountered bad pattern: %s", pattern)
	return false
}

func (c *Context) EmptyEnum(enum []string) bool {
	for _, val := range enum {
		if val == "" {
			return true
		}
	}
	return false
}

func (c *Context) SortedSchemaNames() (names []string) {
	for name := range c.Schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

func (c *Context) ParseSchemas() {
	c.Schemas = make(map[string]*Schema)
	for name, mi := range c.Doc.Schemas {
		s := NewSchema(c, mi, name, nil)
		c.Schemas[name] = s
		s.ParseSubSchemas(c.Schemas)
	}
}

func (c *Context) ParseResources(specs map[string]*googlespec.APIResource, parentName string) []*Resource {
	l := []*Resource{}

	if specs == nil || len(specs) == 0 {
		return l
	}

	resMap := make(map[string]interface{})
	for k, v := range specs {
		resMap[k] = v
	}
	for _, name := range c.SortedKeys(resMap) {
		spec := specs[name]
		r := NewResource(c, name, parentName, spec,
			c.ParseResources(spec.Resources, fmt.Sprintf("%s.%s", parentName, name)))
		l = append(l, r)
	}

	return l
}

func (c *Context) Parse() {
	c.ParseSchemas()

	c.Resources = c.ParseResources(c.Doc.Resources, "")
}

func (c *Context) GenerateService(r *Resource) {
	if (len(r.Methods)) == 0 {
		return
	}

	serviceName := "Service"
	if r != nil {
		serviceName = r.GoType
	}

	//def
	c.Pn("type %s interface{", serviceName)
	for _, m := range r.Methods {
		c.Pn(m.Signature)
	}
	c.Pn("}")
	c.P("\n")

	//default impl
	c.Pn("type Default%s struct{", serviceName)
	c.Pn("}")
	c.Pn("")
	for _, m := range r.Methods {
		c.Pn("func (s *Default%s) %s{", serviceName, m.Signature)
		c.Pn("    return nil,nil")
		c.Pn("}")
		c.Pn("")
	}

	//options
	for _, m := range r.Methods {
		if len(m.OptionalQueryParams) == 0 {
			continue
		}

		c.Pn("type %s struct{", m.OptionParamStructType)
		for _, p := range m.OptionalQueryParams {
			c.Pn("    %s *%s", c.InitialCap(p.GoName), p.GoType)
		}
		c.Pn("}")
		c.Pn("")

		c.Pn("func Parse%s(values url.Values)(_opts *%s,_err error){", m.OptionParamStructType, m.OptionParamStructType)
		c.Pn("    _opts = &%s{}", m.OptionParamStructType)
		c.Pn("")
		c.Pn("    var _str string")
		for _, p := range m.OptionalQueryParams {
			onOptionQueryParamParseError := func() {
				c.Pn("    if _err!=nil{")
				c.Pn("        return nil,_err")
				c.Pn("    }")
				c.Pn("")
				c.Pn("    _opts.%s=&_q", c.InitialCap(p.GoName))
				c.Pn("")
			}

			c.Pn("    _str=values.Get(\"%s\")", p.name)
			c.Pn("    if _str!=\"\"{")
			if p.GoType == "string" {
				c.Pn("    _opts.%s=&_str", c.InitialCap(p.GoName))
				c.Pn("")
			} else if p.GoType == "[]string" {
				c.Pn("    _q,_err:=marsapi.ParseStringList(_str)")
				onOptionQueryParamParseError()
			} else if p.GoType == "bool" {
				c.Pn("    _q,_err:=strconv.ParseBool(_str)")
				onOptionQueryParamParseError()
			} else if p.GoType == "int64" {
				c.Pn("    _q,_err:=strconv.ParseInt(_str,10,64)")
				onOptionQueryParamParseError()
			} else if p.GoType == "[]int64" {
				c.Pn("    _q,_err:=marsapi.ParseInt64List(_str)")
				onOptionQueryParamParseError()
			} else if p.GoType == "uint64" {
				c.Pn("    _q,_err:=strconv.ParseUint(_str,10,64)")
				onOptionQueryParamParseError()
			} else if p.GoType == "[]uint64" {
				c.Pn("    _q,_err:=marsapi.ParseUint64List(_str)")
				onOptionQueryParamParseError()
			} else if p.GoType == "float64" {
				c.Pn("    _q,_err:=strconv.ParseFloat(_str,64)")
				onOptionQueryParamParseError()
			} else if p.GoType == "[]float64" {
				c.Pn("    _q,_err:=marsapi.ParseFloat64List(_str)")
				onOptionQueryParamParseError()
			} else {
				panic("unkown option query param type,meth=" + m.GoName + ",param=" + p.GoName + " " + p.GoType)
			}

			c.Pn("    }")
			c.Pn("")
		}
		c.Pn("    return _opts")
		c.Pn("}")
		c.Pn("")
	}

	//handle
	onHandleError := func() {
		c.Pn("    if _err!=nil{")
		c.Pn("        _ctx.ServiceError=_err")
		c.Pn("        return")
		c.Pn("    }")
		c.Pn("")
	}
	c.Pn("func Handle%s(_r marsapi.Router,_s %s)(err error){", serviceName, serviceName)
	for _, m := range r.Methods {
		//method options
		c.Pn("    %sMethodOptions:=&marsapi.MethodOptions{}", m.GoName)
		if len(m.spec.Scopes) > 0 {
			c.Pn("    %sMethodOptions.Scopes=[]string{\"%s\"}", m.GoName, strings.Join(m.spec.Scopes, "\",\""))
		} else {
			c.Pn("    %sMethodOptions.Scopes=[]string{}", m.GoName)
		}
		if m.spec.SupportsMediaDownload {
			c.Pn("   %sMethodOptions.SupportsMediaDownload=true", m.GoName)
		} else {
			c.Pn("   %sMethodOptions.SupportsMediaDownload=false", m.GoName)
		}
		if m.spec.SupportsMediaUpload {
			c.Pn("   %sMethodOptions.SupportsMediaUpload=true", m.GoName)
		} else {
			c.Pn("   %sMethodOptions.SupportsMediaUpload=false", m.GoName)
		}
		if m.spec.SupportsSubscription {
			c.Pn("   %sMethodOptions.SupportsSubscription=true", m.GoName)
		} else {
			c.Pn("   %sMethodOptions.SupportsSubscription=false", m.GoName)
		}

		//handle func
		c.Pn("    _r.Handle(\"%s\",\"%s\", func(_ctx *marsapi.Context) {", m.spec.HttpMethod, m.spec.Path)
		c.Pn("    var _err error")
		c.Pn("")

		//path params
		for _, p := range m.PathParams {
			valueString := fmt.Sprintf("_ctx.PathParamMap[\"%s\"]", p.name)
			if p.GoType == "string" {
				c.Pn("    _p_%s:=%s", p.GoName, valueString)
				c.Pn("")
			} else if p.GoType == "int64" {
				c.Pn("    _p_%s,_err:=strconv.ParseInt(%s,10,64)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "uint64" {
				c.Pn("    _p_%s,_err:=strconv.ParseUint(%s,10,64)", p.GoName, valueString)
				onHandleError()
			} else {
				panic("unknown path param type,meth=" + m.GoName + ",param=" + p.GoName + " " + p.GoType)
			}
		}

		//query params
		for _, p := range m.RequiredQueryParams {
			valueString := fmt.Sprintf("_ctx.HttpRequest.URL.Query().Get(\"%s\")", p.GoName)
			if p.GoType == "string" {
				c.Pn("    _q_%s:=%s", p.GoName, valueString)
				c.Pn("")
			} else if p.GoType == "[]string" {
				c.Pn("    _q_%s,_err:=marsapi.ParseStringList(%s)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "bool" {
				c.Pn("    _q_%s,_err:=strconv.ParseBool(%s)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "int64" {
				c.Pn("    _q_%s,_err:=strconv.ParseInt(%s,10,64)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "[]int64" {
				c.Pn("    _q_%s,_err:=marsapi.ParseInt64List(%s)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "uint64" {
				c.Pn("    _q_%s,_err:=strconv.ParseUint(%s,10,64)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "[]uint64" {
				c.Pn("    _q_%s,_err:=marsapi.ParseUint64List(%s)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "float64" {
				c.Pn("    _q_%s,_err:=strconv.ParseFloat(%s,64)", p.GoName, valueString)
				onHandleError()
			} else if p.GoType == "[]float64" {
				c.Pn("    _q_%s,_err:=marsapi.ParseFloat64List(%s)", p.GoName, valueString)
				onHandleError()
			} else {
				panic("unkown query param type,meth=" + m.GoName + ",param=" + p.GoName + " " + p.GoType)
			}
		}

		//request body
		if m.RequestType != "" {
			c.Pn("    body,_err:= ioutil.ReadAll(_ctx.HttpRequest.Body)")
			onHandleError()
			c.Pn("    _req:=%s{}", strings.TrimLeft(m.RequestType, "*"))
			c.Pn("    _err=json.Unmarshal(body,&_req)")
			onHandleError()
		}

		//option query params
		if len(m.OptionalQueryParams) > 0 {
			c.Pn("    _opts,_err:=Parse%s(_ctx.HttpRequest.URL.Query())", m.OptionParamStructType)
			c.Pn("    if _err!=nil{")
			c.Pn("        _ctx.ServiceError=_err")
			c.Pn("        return")
			c.Pn("    }")
			c.Pn("")
		}

		//call
		if m.ResponseType != "" {
			c.P("        _resp,_err:=")
		} else {
			c.P("        _err=")
		}
		c.P("_s.%s(_ctx", m.GoName)
		for _, p := range m.PathParams {
			c.P(",_p_%s", p.GoName)
		}
		for _, p := range m.RequiredQueryParams {
			c.P(",_q_%s", p.GoName)
		}
		if m.RequestType != "" {
			c.P(",_req")
		}
		if len(m.OptionalQueryParams) > 0 {
			c.P(",_opts")
		}
		c.Pn(")")
		c.Pn("")
		if m.ResponseType != "" {
			c.Pn("    _ctx.ServiceResponse=_resp")
		}
		c.Pn("    _ctx.ServiceError=_err")
		c.Pn("    },%sMethodOptions)", m.GoName)
		c.Pn("")
	}
	c.Pn("    return nil")
	c.Pn("}")

	if r != nil {
		for _, subResource := range r.resources {
			c.GenerateService(subResource)
		}
	}
}

func GenerateServer(doc_json string, params *ServerGenerateParams) (code string, err error) {
	panic("111")

	doc := &googlespec.APIDocument{}
	err = json.Unmarshal([]byte(doc_json), doc)
	if err != nil {
		return "", err
	}

	c := Context{}
	c.Namespace = params.Namespace

	c.Code = &bytes.Buffer{}
	c.Doc = doc

	c.Parse()

	c.Pn("package %s", c.Namespace)
	c.P("\n")

	c.Pn("import \"io/ioutil\"")
	c.Pn("import \"encoding/json\"")
	c.Pn("import \"errors\"")
	c.Pn("import \"net/http\"")
	c.Pn("import \"net/url\"")
	c.Pn("import \"strconv\"")
	c.Pn("import \"github.com/marshome/apis/marsapi\"")
	c.P("\n")

	c.Pn("var _=errors.New(\"\")")
	c.Pn("var _=http.DefaultClient")
	c.Pn("var _=&url.URL{}")
	c.Pn("var _=strconv.ErrRange")
	c.Pn("var _=ioutil.Discard")
	c.Pn("var _=json.InvalidUTF8Error{}")

	for _, name := range c.SortedSchemaNames() {
		c.Schemas[name].GenerateSchema()
	}

	for _, r := range c.Resources {
		c.GenerateService(r)
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}
