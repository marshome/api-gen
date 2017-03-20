package generate

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"errors"

	"github.com/marshome/apis/spec"
)

var (
	BaseURL = flag.String("base_url", "", "(optional) Override the default service API URL. If empty, the service's root URL will be used.")
)

type ServerGenerateParams struct {
	Namespace string
}

type Context struct {
	ApiPackageBase string
	Namespace string

	Doc            *spec.APIDocument
	Code           *bytes.Buffer

	usedNames     namePool
	ResponseTypes map[string]bool
	RequestTypes  map[string]bool

	Schemas    map[string]*Schema
	Resources  []*Resource
	APIMethods []*Method
}

// convertMultiParams builds a []string temp variable from a slice
// of non-strings and returns the name of the temp variable.
func (c *Context) ConvertMultiParams(param string) string {
	c.Pn(" var %v_ []string", param)
	c.Pn(" for _, v := range %v {", param)
	c.Pn("  %v_ = append(%v_, fmt.Sprint(v))", param, param)
	c.Pn(" }")
	return param + "_"
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
	s := fmt.Sprintf(format, args...)
	c.Code.WriteString(s)
}

func (c *Context) Pn(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	c.Code.WriteString(s)
	c.Code.WriteString("\n")
}

func (c *Context) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (c *Context) Package() (pkg string) {
	return strings.ToLower(c.Doc.Name)
}

func (c *Context) Target() (target string) {
	return fmt.Sprintf("%s/%s/%s", c.ApiPackageBase, c.Package(), c.RenameVersion(c.Doc.Version))
}

// oddVersionRE matches unusual API names like directory_v1.
var oddVersionRE = regexp.MustCompile(`^(.+)_(v[\d\.]+)$`)

// renameVersion conditionally rewrites the provided version such
// that the final path component of the import path doesn't look
// like a Go identifier. This keeps the consistency that import paths
// for the generated Go packages look like:
//     google.golang.org/api/NAME/v<version>
// and have package NAME.
// See https://github.com/google/google-api-go-client/issues/78
func (c *Context) RenameVersion(version string) string {
	if version == "alpha" || version == "beta" {
		return "v0." + version
	}
	if m := oddVersionRE.FindStringSubmatch(version); m != nil {
		return m[1] + "/" + m[2]
	}
	return version
}

func (c *Context) ResolveRelative(basestr, relstr string) string {
	u, err := url.Parse(basestr)
	if err != nil {
		c.Panicf("Error parsing base URL %q: %v", basestr, err)
	}
	rel, err := url.Parse(relstr)
	if err != nil {
		c.Panicf("Error parsing relative URL %q: %v", relstr, err)
	}
	u = u.ResolveReference(rel)
	return u.String()
}

func (c *Context) ApiBaseURL() string {
	var base, rel string
	switch {
	case *BaseURL != "":
		//base, rel = *baseURL, jstr(a.m, "basePath")//todo
	case c.Doc.RootUrl != "":
		base, rel = c.Doc.RootUrl, c.Doc.ServicePath
	default:
		//base, rel = *apisURL, jstr(a.m, "basePath")
	}
	return c.ResolveRelative(base, rel)
}

func (c *Context) SortedKeys(m map[string]interface{}) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}

func (c *Context) ValidGoIdentifer(ident string) string {
	id := Depunct(ident, false)
	switch id {
	case "break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var":
		return id + "_"
	}
	return id
}

func (c *Context) ScopeIdentifierFromURL(urlStr string) string {
	const prefix = "https://www.googleapis.com/auth/"
	if !strings.HasPrefix(urlStr, prefix) {
		const https = "https://"
		if !strings.HasPrefix(urlStr, https) {
			log.Fatalf("Unexpected oauth2 scope %q doesn't start with %q", urlStr, https)
		}
		ident := c.ValidGoIdentifer(Depunct(urlStr[len(https):], true)) + "Scope"
		return ident
	}
	ident := c.ValidGoIdentifer(c.InitialCap(urlStr[len(prefix):])) + "Scope"
	return ident
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

func (c *Context) GenerateScopeConstants() {
	if c.Doc.Auth == nil {
		return
	}

	if c.Doc.Auth.OAuth2 == nil {
		return
	}

	scopes := c.Doc.Auth.OAuth2.Scopes
	if scopes == nil || len(scopes) == 0 {
		return
	}

	scopes_interfaces := make(map[string]interface{})
	for k, v := range scopes {
		scopes_interfaces[k] = v
	}

	c.Pn("// OAuth2 scopes used by this API.")
	c.Pn("const (")
	n := 0
	for _, scopeName := range c.SortedKeys(scopes_interfaces) {
		mi := scopes[scopeName]
		if n > 0 {
			c.P("\n")
		}
		n++
		ident := c.ScopeIdentifierFromURL(scopeName)
		if des := mi.Description; des != "" {
			c.P("%s", c.AsComment("\t", des))
		}
		c.Pn("\t%s = %q", ident, scopeName)
	}
	c.P(")\n\n")
}

func (c *Context) AddFieldValueComments(p func(format string, args ...interface{}), field Field, indent string, blankLine bool) {
	var lines []string

	if field.Enum() != nil {
		desc := field.EnumDescriptions()
		lines = append(lines, c.AsComment(indent, "Possible values:"))
		defval := field.Default()
		for i, v := range field.Enum() {
			more := ""
			if v == defval {
				more = " (default)"
			}
			if len(desc) > i && desc[i] != "" {
				more = more + " - " + desc[i]
			}
			lines = append(lines, c.AsComment(indent, `  "`+v+`"`+more))
		}
	} else if field.UnfortunateDefault() {
		lines = append(lines, c.AsComment("\t", fmt.Sprintf("Default: %s", field.Default())))
	}
	if blankLine && len(lines) > 0 {
		p(indent + "//\n")
	}
	for _, l := range lines {
		p("%s", l)
	}
}

func (c *Context) NeedsDataWrapper() bool {
	if c.Doc.Features == nil {
		return false
	}

	for _, feature := range c.Doc.Features {
		if feature == "dataWrapper" {
			return true
		}
	}
	return false
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

func (c *Context) MustSimpleTypeConvert(apiType, format string) string {
	if gotype, ok := c.SimpleTypeConvert(apiType, format); ok {
		return gotype
	}
	panic(fmt.Sprintf("failed to simpleTypeConvert(%q, %q)", apiType, format))
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

func (c *Context) ParseSchemas() {
	if c.Schemas != nil {
		panic("")
	}
	c.Schemas = make(map[string]*Schema)
	for name, mi := range c.Doc.Schemas {
		s := NewSchema(c, mi, name, nil)
		c.Schemas[name] = s
		err := s.populateSubSchemas()
		if err != nil {
			c.Panicf("Error populating schema with API name %q: %v", name, err)
		}
	}
}

func (c *Context) ParseResources(rm map[string]*spec.APIResource, p string) []*Resource {
	res := []*Resource{}

	if rm == nil || len(rm) == 0 {
		return res
	}

	resMap := make(map[string]interface{})
	for k, v := range rm {
		resMap[k] = v
	}
	for _, rname := range c.SortedKeys(resMap) {
		r := rm[rname]
		res = append(res, &Resource{c, rname, p, r, c.ParseResources(r.Resources, fmt.Sprintf("%s.%s", p, rname)), nil})
	}
	return res
}

func (c *Context) ParseAPIMethods() []*Method {
	meths := []*Method{}
	methMap := make(map[string]interface{})
	for k, v := range c.Doc.Methods {
		methMap[k] = v
	}
	for _, name := range c.SortedKeys(methMap) {
		mi := c.Doc.Methods[name]
		meths = append(meths, &Method{
			c:    c,
			r:    nil, // to be explicit
			name: name,
			doc:  mi,
		})
	}
	return meths
}

func (c *Context) SortedSchemaNames() (names []string) {
	for name := range c.Schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

func (c *Context) Parse() {
	c.ParseSchemas()

	c.APIMethods = c.ParseAPIMethods()

	c.Resources = c.ParseResources(c.Doc.Resources, "")

	for _, r := range c.Resources {
		r.ParseMethods()
	}

	c.ResponseTypes = make(map[string]bool)
	c.RequestTypes = make(map[string]bool)
	for _, meth := range c.APIMethods {
		meth.CacheRequestTypes()
		meth.CacheResponseTypes()
	}
	for _, res := range c.Resources {
		res.CacheRequestTypes()
		res.CacheResponseTypes()
	}
}

func (c *Context) GenerateService(r *Resource, methods []*Method) {
	if (len(methods)) == 0 {
		return
	}

	serviceName := "Service"
	if r != nil {
		serviceName = r.GoType()
	}

	//options
	for _, meth := range methods {
		if len(meth.OptParams()) > 0 {
			c.Pn("type %s struct{", meth.OptionsType())
			c.Pn("}")
			c.Pn("")
		}
	}

	//def
	c.Pn("type %s interface{", serviceName)
	for _, meth := range methods {
		c.Pn(meth.Signature())
	}
	c.Pn("}")
	c.P("\n")

	//default impl
	//c.Pn("type Default%s struct{", serviceName)
	//c.Pn("}")
	//c.Pn("")
	//for _, m := range methods {
	//	c.Pn("func (s *Default%s) %s{", serviceName, m.Signature())
	//	c.Pn("    return nil,nil")
	//	c.Pn("}")
	//	c.Pn("")
	//}

	//handle
	c.Pn("func Handle%s(r marsapi.Router,s %s)(err error){", serviceName, serviceName)
	for _, m := range methods {
		c.Pn("    r.Handle(\"%s\",\"%s\", func(ctx *marsapi.Context) {", m.doc.HttpMethod, m.doc.Path)
		for _, param := range m.NewArguments().l {
			if param.location == "path" {
			} else if param.location == "query" {

			} else if param.location == "body" {
			} else {
				panic(errors.New("unkown location"))
			}
		}

		//c.Pn("        s.%s(ctx)", m.GoName())
		c.Pn("    })")
		c.Pn("")
	}
	c.Pn("    return nil")
	c.Pn("}")

	if r != nil {
		for _, subResource := range r.resources {
			c.GenerateService(subResource, methods)
		}
	}
}

func GenerateServer(doc_json string, params *ServerGenerateParams) (code string, err error) {
	doc := &spec.APIDocument{}
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

	c.Pn("import \"errors\"")
	c.Pn("import \"net/http\"")
	c.Pn("import \"github.com/marshome/apis/marsapi\"")
	c.P("\n")

	c.Pn("var _=errors.New(\"\")")
	c.Pn("var _=http.DefaultClient")

	for _, name := range c.SortedSchemaNames() {
		c.Schemas[name].WriteSchemaCode()
	}

	c.GenerateService(nil, c.APIMethods)

	for _, r := range c.Resources {
		c.GenerateService(r, r.Methods)
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}
