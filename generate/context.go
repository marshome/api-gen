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

	"github.com/marshome/apis/spec"
)

const googleDiscoveryURL = "https://www.googleapis.com/discovery/v1/apis"

var (
	BaseURL = flag.String("base_url", "", "(optional) Override the default service API URL. If empty, the service's root URL will be used.")
	ApisURL = flag.String("discoveryurl", googleDiscoveryURL, "URL to root discovery document")

	ContextHTTPPkg = flag.String("ctxhttp_pkg", "golang.org/x/net/context/ctxhttp", "Go package path of the 'ctxhttp' package.")
	ContextPkg     = flag.String("context_pkg", "golang.org/x/net/context", "Go package path of the 'context' package.")
	GensupportPkg  = flag.String("gensupport_pkg", "airble.com/pkg/gensupport", "Go package path of the 'api/gensupport' support package.")
	GoogleapiPkg   = flag.String("googleapi_pkg", "google.golang.org/api/googleapi", "Go package path of the 'api/googleapi' support package.")
)

var canonicalDocsURL = map[string]string{}

type ClientGenerateParams struct {
	ApiPackageBase string
}

type ServerGenerateParams struct {
	ApiPackageBase string
}

type Context struct {
	ApiPackageBase string
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

// emptyPattern reports whether a pattern matches the empty string.
func (c *Context) EmptyPattern(pattern string) bool {
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString("")
	}
	log.Printf("Encountered bad pattern: %s", pattern)
	return false
}

// emptyEnum reports whether a property enum list contains the empty string.
func (c *Context) EmptyEnum(enum []string) bool {
	for _, val := range enum {
		if val == "" {
			return true
		}
	}
	return false
}

// PopulateSchemas reads all the API types ("schemas") from the JSON file
// and converts them to *Schema instances, returning an identically
// keyed map, additionally containing subresources.  For instance,
//
// A resource "Foo" of type "object" with a property "bar", also of type
// "object" (an anonymous sub-resource), will get a synthetic API name
// of "Foo.bar".
//
// A resource "Foo" of type "array" with an "items" of type "object"
// will get a synthetic API name of "Foo.Item".
func (c *Context) ParseSchemas() {
	if c.Schemas != nil {
		panic("")
	}
	c.Schemas = make(map[string]*Schema)
	for name, mi := range c.Doc.Schemas {
		s := &Schema{
			c:       c,
			apiName: name,
			m:       mi,
		}

		// And a little gross hack, so a map alone is good
		// enough to get its apiName:
		s._apiName = name

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

// APIMethods returns top-level ("API-level") methods. They don't have an associated resource.
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

func GenerateClient(doc_json string, params *ClientGenerateParams) (code string, err error) {
	doc := &spec.APIDocument{}
	err = json.Unmarshal([]byte(doc_json), doc)
	if err != nil {
		return "", err
	}

	c := Context{}
	c.Code = &bytes.Buffer{}
	c.Doc = doc
	c.ApiPackageBase = params.ApiPackageBase

	c.Parse()

	c.Pn("// Package %s provides access to the %s.", c.Package(), c.Doc.Title)
	if c.Doc.DocumentationLink != "" {
		c.Pn("//")
		c.Pn("// See %s", c.Doc.DocumentationLink)
	}
	c.Pn("//\n// Usage example:")
	c.Pn("//")
	c.Pn("//   import %q", c.Target())
	c.Pn("//   ...")
	c.Pn("//   %sService, err := %s.New(oauthHttpClient)", c.Package(), c.Package())

	c.Pn("package %s //// import %q", c.Package(), c.Target())
	c.P("\n")
	c.Pn("import (")
	for _, imp := range []struct {
		pkg   string
		lname string
	}{
		{"bytes", ""},
		{"encoding/json", ""},
		{"errors", ""},
		{"fmt", ""},
		{"io", ""},
		{"net/http", ""},
		{"net/url", ""},
		{"strconv", ""},
		{"strings", ""},
		{*ContextHTTPPkg, "ctxhttp"},
		{*ContextPkg, "context"},
		{*GensupportPkg, "gensupport"},
		{*GoogleapiPkg, "googleapi"},
	} {
		if imp.lname == "" {
			c.Pn("  %q", imp.pkg)
		} else {
			c.Pn("  %s %q", imp.lname, imp.pkg)
		}
	}
	c.Pn(")")
	c.Pn("\n// Always reference these packages, just in case the auto-generated code")
	c.Pn("// below doesn't.")
	c.Pn("var _ = bytes.NewBuffer")
	c.Pn("var _ = strconv.Itoa")
	c.Pn("var _ = fmt.Sprintf")
	c.Pn("var _ = json.NewDecoder")
	c.Pn("var _ = io.Copy")
	c.Pn("var _ = url.Parse")
	c.Pn("var _ = gensupport.MarshalJSON")
	c.Pn("var _ = googleapi.Version")
	c.Pn("var _ = errors.New")
	c.Pn("var _ = strings.Replace")
	c.Pn("var _ = context.Canceled")
	c.Pn("var _ = ctxhttp.Do")
	c.Pn("")
	c.Pn("const apiId = %q", c.Doc.Id)
	c.Pn("const apiName = %q", c.Doc.Name)
	c.Pn("const apiVersion = %q", c.Doc.Version)
	c.Pn("const basePath = %q", c.ApiBaseURL())

	c.GenerateScopeConstants()

	c.GetName("New") // ignore return value; we're the first caller
	c.Pn("func New(client *http.Client) (*Service, error) {")
	c.Pn("if client == nil { return nil, errors.New(\"client is nil\") }")
	c.Pn("s := &Service{client: client, BasePath: basePath}")
	for _, res := range c.Resources { // add top level resources.
		c.Pn("s.%s = New%s(s)", res.GoField(), res.GoType())
	}
	c.Pn("return s, nil")
	c.Pn("}")

	c.GetName("Service") // ignore return value; no user-defined names yet
	c.Pn("\ntype Service struct {")
	c.Pn(" client *http.Client")
	c.Pn(" BasePath string // API endpoint base URL")
	c.Pn(" UserAgent string // optional additional User-Agent fragment")

	for _, res := range c.Resources {
		c.Pn("\n\t%s\t*%s", res.GoField(), res.GoType())
	}
	c.Pn("}")
	c.Pn("\nfunc (s *Service) userAgent() string {")
	c.Pn(` if s.UserAgent == "" { return googleapi.UserAgent }`)
	c.Pn(` return googleapi.UserAgent + " " + s.UserAgent`)
	c.Pn("}\n")

	for _, res := range c.Resources {
		res.GenerateType()
	}

	for _, name := range c.SortedSchemaNames() {
		c.Schemas[name].WriteSchemaCode()
	}

	for _, meth := range c.APIMethods {
		meth.GenerateClientCode()
	}

	for _, res := range c.Resources {
		res.GenerateClientMethods()
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}

func (c *Context) GenerateServerService(r *Resource, methods []*Method) {
	path_params_buffer := &bytes.Buffer{}
	regist_service_buffer := &bytes.Buffer{}

	if r == nil {
		c.Pn("type Service interface{")
	} else {
		c.Pn("type " + r.GoType() + " interface{")
	}

	for _, meth := range methods {
		meth.GenerateServerCode(path_params_buffer, regist_service_buffer)
	}
	c.Pn("}")
	c.P("\n")

	c.Pn(path_params_buffer.String())
	c.P("\n")

	if len(methods) > 0 {
		if r == nil {
			c.Pn("func RegistServiceService(service Service)(err error){")
		} else {
			c.Pn("func Regist" + r.GoType() + "(service " + r.GoType() + ")(err error){")
		}
		if r == nil {
			c.Pn("s,err:=endpoints.RegisterService(service,\"\",\"\",\"" + c.Doc.Version + "\",\"" + c.Doc.Title + "\",true)")
		} else {
			c.Pn("s,err:=endpoints.RegisterService(service,\"" + r.parent + "\",\"" + r.name + "\",\"" + c.Doc.Version + "\",\"" + c.Doc.Title + "\",true)")
		}
		c.Pn("if err!=nil{")
		c.Pn("return err")
		c.Pn("}")
		c.P("\n")
		c.P("var m *endpoints.ServiceMethod")
		c.P("\n")
		c.P(regist_service_buffer.String())
		c.P("return nil")
		c.Pn("}")
	}
}

func GenerateServer(doc_json string, params *ServerGenerateParams) (code string, err error) {
	doc := &spec.APIDocument{}
	err = json.Unmarshal([]byte(doc_json), doc)
	if err != nil {
		return "", err
	}

	c := Context{}
	c.Code = &bytes.Buffer{}
	c.Doc = doc
	c.ApiPackageBase = params.ApiPackageBase

	c.Parse()

	c.Pn("// Package %s provides access to the %s.", c.Package(), c.Doc.Title)
	if c.Doc.DocumentationLink != "" {
		c.Pn("//")
		c.Pn("// See %s", c.Doc.DocumentationLink)
	}
	c.Pn("//\n// Usage example:")
	c.Pn("//")
	c.Pn("//   import %q", c.Target())
	c.Pn("//   ...")
	c.Pn("//   %sService, err := %s.New(oauthHttpClient)", c.Package(), c.Package())

	c.Pn("package api")
	c.P("\n")

	c.Pn("import \"airble.com/services/apis/endpoints\"")
	c.Pn("import \"airble.com/pkg/gensupport\"")
	c.Pn("import \"errors\"")
	c.Pn("import \"golang.org/x/net/context\"")
	c.Pn("import googleapi \"google.golang.org/api/googleapi\"")
	c.P("\n")

	for _, name := range c.SortedSchemaNames() {
		c.Schemas[name].WriteSchemaCode()
	}

	c.GenerateServerService(nil, c.APIMethods)

	for _, res := range c.Resources {
		res.GenerateServerMethods()
	}

	clean, err := format.Source(c.Code.Bytes())
	if err != nil {
		return c.Code.String(), err
	}
	return string(clean), nil
}
