package generate

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/marshome/apis/spec"
)

type Method struct {
	c    *Context
	r    *Resource // or nil if a API-level (top-level) method
	name string

	doc *spec.APIMethod

	params []*Param // all Params, of each type, lazily set by first access to Parameters

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

func (meth *Method)GetRequestTypeReal()(typ string) {
	reqType := meth.GetRequestType()
	if reqType != "" {
		return strings.TrimLeft(reqType, "*")
	}

	if meth.r == nil {
		return "Service_" + meth.c.InitialCap(meth.name)+"Request"
	} else {
		return "" + meth.c.InitialCap(meth.r.parent + meth.r.name) + "Service_" + meth.c.InitialCap(meth.name)+"Request"
	}

	return ""
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

func (meth *Method) Signature()(signature string) {
	signature+="    " + meth.GoName() + "(ctx *restful.Context"

	reqType := meth.GetRequestTypeReal()
	if reqType != "" {
		signature+=", req *" + reqType
	}

	signature+=")"

	resType := meth.GetResponseType()
	if resType == "" {
		signature+="(err error)"
	} else {
		signature+="(resp " + resType + ", err error)"
	}

	return signature
}

func (meth *Method) GenerateClientCode() {
	res := meth.r // may be nil if a top-level method
	a := meth.c
	p, pn := a.P, a.Pn

	pn("\n// method id %q:", meth.Id())

	retType := meth.GetResponseType()
	retTypeComma := retType
	if retTypeComma != "" {
		retTypeComma += ", "
	}

	args := meth.NewArguments()
	methodName := a.InitialCap(meth.name)
	prefix := ""
	if res != nil {
		prefix = a.InitialCap(fmt.Sprintf("%s.%s", res.parent, res.name))
	}
	callName := a.GetName(prefix + methodName + "Call")

	pn("\ntype %s struct {", callName)
	pn(" s *Service")
	for _, arg := range args.l {
		if arg.location != "query" {
			pn(" %s %s", arg.goname, arg.gotype)
		}
	}
	pn(" urlParams_ generate.URLParams")
	httpMethod := meth.doc.HttpMethod
	if httpMethod == "GET" {
		pn(" ifNoneMatch_ string")
	}
	if meth.supportsMediaUpload() {
		pn(" media_     io.Reader")
		pn(" resumable_ googleapi.SizeReaderAt")
		pn(" mediaType_ string")
		pn(" protocol_  string")
		pn(" progressUpdater_  googleapi.ProgressUpdater")
	}
	pn(" ctx_ context.Context")
	pn("}")

	p("\n%s", a.AsComment("", methodName+": "+meth.doc.Description))
	if res != nil {
		if url := canonicalDocsURL[fmt.Sprintf("%v%v/%v", meth.c.Doc.DocumentationLink, res.name, meth.name)]; url != "" {
			pn("// For details, see %v", url)
		}
	}

	var servicePtr string
	if res == nil {
		pn("func (s *Service) %s(%s) *%s {", methodName, args, callName)
		servicePtr = "s"
	} else {
		pn("func (r *%s) %s(%s) *%s {", res.GoType(), methodName, args, callName)
		servicePtr = "r.s"
	}

	pn(" c := &%s{s: %s, urlParams_: make(generate.URLParams)}", callName, servicePtr)
	for _, arg := range args.l {
		// TODO(gmlewis): clean up and consolidate this section.
		// See: https://code-review.googlesource.com/#/c/3520/18/google-api-go-generator/gen.go
		if arg.location == "query" {
			switch arg.gotype {
			case "[]string":
				pn(" c.urlParams_.SetMulti(%q, append([]string{}, %v...))", arg.apiname, arg.goname)
			case "string":
				pn(" c.urlParams_.Set(%q, %v)", arg.apiname, arg.goname)
			default:
				if strings.HasPrefix(arg.gotype, "[]") {
					tmpVar := a.ConvertMultiParams(arg.goname)
					pn(" c.urlParams_.SetMulti(%q, %v)", arg.apiname, tmpVar)
				} else {
					pn(" c.urlParams_.Set(%q, fmt.Sprint(%v))", arg.apiname, arg.goname)
				}
			}
			continue
		}
		if arg.gotype == "[]string" {
			pn(" c.%s = append([]string{}, %s...)", arg.goname, arg.goname) // Make a copy of the []string.
			continue
		}
		pn(" c.%s = %s", arg.goname, arg.goname)
	}
	pn(" return c")
	pn("}")

	for _, opt := range meth.OptParams() {
		if opt.Location() != "query" {
			a.Panicf("optional parameter has unsupported location %q", opt.Location())
		}
		setter := a.InitialCap(opt.name)
		des := meth.doc.Description
		des = strings.Replace(des, "Optional.", "", 1)
		des = strings.TrimSpace(des)
		p("\n%s", a.AsComment("", fmt.Sprintf("%s sets the optional parameter %q: %s", setter, opt.name, des)))
		a.AddFieldValueComments(p, opt, "", true)
		np := new(namePool)
		np.Get("c") // take the receiver's name
		paramName := np.Get(a.ValidGoIdentifer(opt.name))
		typePrefix := ""
		if opt.IsRepeated() {
			typePrefix = "..."
		}
		pn("func (c *%s) %s(%s %s%s) *%s {", callName, setter, paramName, typePrefix, opt.GoType(), callName)
		if opt.IsRepeated() {
			if opt.GoType() == "string" {
				pn("c.urlParams_.SetMulti(%q, append([]string{}, %v...))", opt.name, paramName)
			} else {
				tmpVar := a.ConvertMultiParams(paramName)
				pn(" c.urlParams_.SetMulti(%q, %v)", opt.name, tmpVar)
			}
		} else {
			if opt.GoType() == "string" {
				pn("c.urlParams_.Set(%q, %v)", opt.name, paramName)
			} else {
				pn("c.urlParams_.Set(%q, fmt.Sprint(%v))", opt.name, paramName)
			}
		}
		pn("return c")
		pn("}")
	}

	if meth.supportsMediaUpload() {
		comment := "Media specifies the media to upload in a single chunk. " +
			"At most one of Media and ResumableMedia may be set."
		p("\n%s", a.AsComment("", comment))
		pn("func (c *%s) Media(r io.Reader) *%s {", callName, callName)
		pn("c.media_ = r")
		pn(`c.protocol_ = "multipart"`)
		pn("return c")
		pn("}")
		comment = "ResumableMedia specifies the media to upload in chunks and can be canceled with ctx. " +
			"At most one of Media and ResumableMedia may be set. " +
			`mediaType identifies the MIME media type of the upload, such as "image/png". ` +
			`If mediaType is "", it will be auto-detected. ` +
			`The provided ctx will supersede any context previously provided to ` +
			`the Context method.`
		p("\n%s", a.AsComment("", comment))
		pn("func (c *%s) ResumableMedia(ctx context.Context, r io.ReaderAt, size int64, mediaType string) *%s {", callName, callName)
		pn("c.ctx_ = ctx")
		pn("c.resumable_ = io.NewSectionReader(r, 0, size)")
		pn("c.mediaType_ = mediaType")
		pn(`c.protocol_ = "resumable"`)
		pn("return c")
		pn("}")
		comment = "ProgressUpdater provides a callback function that will be called after every chunk. " +
			"It should be a low-latency function in order to not slow down the upload operation. " +
			"This should only be called when using ResumableMedia (as opposed to Media)."
		p("\n%s", a.AsComment("", comment))
		pn("func (c *%s) ProgressUpdater(pu googleapi.ProgressUpdater) *%s {", callName, callName)
		pn(`c.progressUpdater_ = pu`)
		pn("return c")
		pn("}")
	}

	comment := "Fields allows partial responses to be retrieved. " +
		"See https://developers.google.com/gdata/docs/2.0/basics#PartialResponse " +
		"for more information."
	p("\n%s", a.AsComment("", comment))
	pn("func (c *%s) Fields(s ...googleapi.Field) *%s {", callName, callName)
	pn(`c.urlParams_.Set("fields", googleapi.CombineFields(s))`)
	pn("return c")
	pn("}")
	if httpMethod == "GET" {
		// Note that non-GET responses are excluded from supporting If-None-Match.
		// See https://github.com/google/google-api-go-client/issues/107 for more info.
		comment := "IfNoneMatch sets the optional parameter which makes the operation fail if " +
			"the object's ETag matches the given value. This is useful for getting updates " +
			"only after the object has changed since the last request. " +
			"Use googleapi.IsNotModified to check whether the response error from Do " +
			"is the result of In-None-Match."
		p("\n%s", a.AsComment("", comment))
		pn("func (c *%s) IfNoneMatch(entityTag string) *%s {", callName, callName)
		pn(" c.ifNoneMatch_ = entityTag")
		pn(" return c")
		pn("}")
	}

	doMethod := "Do method"
	if meth.supportsMediaDownload() {
		doMethod = "Do and Download methods"
	}
	commentFmtStr := "Context sets the context to be used in this call's %s. " +
		"Any pending HTTP request will be aborted if the provided context is canceled."
	comment = fmt.Sprintf(commentFmtStr, doMethod)
	p("\n%s", a.AsComment("", comment))
	if meth.supportsMediaUpload() {
		comment = "This context will supersede any context previously provided to " +
			"the ResumableMedia method."
		p("%s", a.AsComment("", comment))
	}
	pn("func (c *%s) Context(ctx context.Context) *%s {", callName, callName)
	pn(`c.ctx_ = ctx`)
	pn("return c")
	pn("}")

	pn("\nfunc (c *%s) doRequest(alt string) (*http.Response, error) {", callName)
	pn("var body io.Reader = nil")
	hasContentType := false
	if ba := args.BodyArg(); ba != nil && httpMethod != "GET" {
		style := "WithoutDataWrapper"
		if a.NeedsDataWrapper() {
			style = "WithDataWrapper"
		}
		pn("body, err := googleapi.%s.JSONReader(c.%s)", style, ba.goname)
		pn("if err != nil { return nil, err }")
		pn(`ctype := "application/json"`)
		hasContentType = true
	}
	pn(`c.urlParams_.Set("alt", alt)`)

	pn("urls := googleapi.ResolveRelative(c.s.BasePath, %q)", meth.doc.Path)
	if meth.supportsMediaUpload() {
		pn("if c.media_ != nil || c.resumable_ != nil {")
		// Hack guess, since we get a 404 otherwise:
		//pn("urls = googleapi.ResolveRelative(%q, %q)", a.apiBaseURL(), meth.mediaUploadPath())
		// Further hack.  Discovery doc is wrong?
		pn("urls = strings.Replace(urls, %q, %q, 1)", "https://www.googleapis.com/", "https://www.googleapis.com/upload/")
		pn(`c.urlParams_.Set("uploadType", c.protocol_)`)
		pn("}")
	}
	pn("urls += \"?\" + c.urlParams_.Encode()")
	if meth.supportsMediaUpload() && httpMethod != "GET" {
		if !hasContentType { // Support mediaUpload but no ctype set.
			pn("body = new(bytes.Buffer)")
			pn(`ctype := "application/json"`)
			hasContentType = true
		}
		pn(`if c.protocol_ != "resumable" {`)
		pn(`  var cancel func()`)
		pn("  cancel, _ = googleapi.ConditionallyIncludeMedia(c.media_, &body, &ctype)")
		pn("  if cancel != nil { defer cancel() }")
		pn("}")
	}
	pn("req, _ := http.NewRequest(%q, urls, body)", httpMethod)
	// Replace param values after NewRequest to avoid reencoding them.
	// E.g. Cloud Storage API requires '%2F' in entity param to be kept, but url.Parse replaces it with '/'.
	argsForLocation := args.ForLocation("path")
	if len(argsForLocation) > 0 {
		pn(`googleapi.Expand(req.URL, map[string]string{`)
		for _, arg := range argsForLocation {
			pn(`"%s": %s,`, arg.apiname, arg.ExprAsString("c."))
		}
		pn(`})`)
	} else {
		// Just call SetOpaque since we aren't calling Expand
		pn(`googleapi.SetOpaque(req.URL)`)
	}

	if meth.supportsMediaUpload() {
		pn(`if c.protocol_ == "resumable" {`)
		pn(` if c.mediaType_ == "" {`)
		pn("  c.mediaType_ = googleapi.DetectMediaType(c.resumable_)")
		pn(" }")
		pn(` req.Header.Set("X-Upload-Content-Type", c.mediaType_)`)
		pn(` req.Header.Set("Content-Type", "application/json; charset=utf-8")`)
		pn("} else {")
		pn(` req.Header.Set("Content-Type", ctype)`)
		pn("}")
	} else if hasContentType {
		pn(`req.Header.Set("Content-Type", ctype)`)
	}
	pn(`req.Header.Set("User-Agent", c.s.userAgent())`)
	if httpMethod == "GET" {
		pn(`if c.ifNoneMatch_ != "" {`)
		pn(` req.Header.Set("If-None-Match", c.ifNoneMatch_)`)
		pn("}")
	}
	pn("if c.ctx_ != nil {")
	pn(" return ctxhttp.Do(c.ctx_, c.s.client, req)")
	pn("}")
	pn("return c.s.client.Do(req)")
	pn("}")

	if meth.supportsMediaDownload() {
		pn("\n// Download fetches the API endpoint's \"media\" value, instead of the normal")
		pn("// API response value. If the returned error is nil, the Response is guaranteed to")
		pn("// have a 2xx status code. Callers must close the Response.Body as usual.")
		pn("func (c *%s) Download() (*http.Response, error) {", callName)
		pn(`res, err := c.doRequest("media")`)
		pn("if err != nil { return nil, err }")
		pn("if err := googleapi.CheckMediaResponse(res); err != nil {")
		pn("res.Body.Close()")
		pn("return nil, err")
		pn("}")
		pn("return res, nil")
		pn("}")
	}

	mapRetType := strings.HasPrefix(retTypeComma, "map[")
	pn("\n// Do executes the %q call.", meth.doc.Id)
	if retTypeComma != "" && !mapRetType {
		commentFmtStr := "Exactly one of %v or error will be non-nil. " +
			"Any non-2xx status code is an error. " +
			"Response headers are in either %v.ServerResponse.Header " +
			"or (if a response was returned at all) in error.(*googleapi.Error).Header. " +
			"Use googleapi.IsNotModified to check whether the returned error was because " +
			"http.StatusNotModified was returned."
		comment := fmt.Sprintf(commentFmtStr, retType, retType)
		p("%s", a.AsComment("", comment))
	}
	pn("func (c *%s) Do() (%serror) {", callName, retTypeComma)
	nilRet := ""
	if retTypeComma != "" {
		nilRet = "nil, "
	}
	pn(`res, err := c.doRequest("json")`)
	if retTypeComma != "" && !mapRetType {
		pn("if res != nil && res.StatusCode == http.StatusNotModified {")
		pn(" if res.Body != nil { res.Body.Close() }")
		pn(" return nil, &googleapi.Error{")
		pn("  Code: res.StatusCode,")
		pn("  Header: res.Header,")
		pn(" }")
		pn("}")
	}
	pn("if err != nil { return %serr }", nilRet)
	pn("defer googleapi.CloseBody(res)")
	pn("if err := googleapi.CheckResponse(res); err != nil { return %serr }", nilRet)
	if meth.supportsMediaUpload() {
		pn(`if c.protocol_ == "resumable" {`)
		pn(` loc := res.Header.Get("Location")`)
		pn(" rx := &googleapi.ResumableUpload{")
		pn("  Client:        c.s.client,")
		pn("  UserAgent:     c.s.userAgent(),")
		pn("  URI:           loc,")
		pn("  Media:         c.resumable_,")
		pn("  MediaType:     c.mediaType_,")
		pn("  ContentLength: c.resumable_.Size(),")
		pn("  Callback:      c.progressUpdater_,")
		pn(" }")
		pn(" res, err = rx.Upload(c.ctx_)")
		pn(" if err != nil { return %serr }", nilRet)
		pn(" defer res.Body.Close()")
		pn("}")
	}
	if retTypeComma == "" {
		pn("return nil")
	} else {
		if mapRetType {
			pn("var ret %s", meth.GetResponseType())
		} else {
			pn("ret := &%s{", meth.ResponseTypeLiteral())
			pn(" ServerResponse: googleapi.ServerResponse{")
			pn("  Header: res.Header,")
			pn("  HTTPStatusCode: res.StatusCode,")
			pn(" },")
			pn("}")
		}
		pn("if err := json.NewDecoder(res.Body).Decode(&ret); err != nil { return nil, err }")
		pn("return ret, nil")
	}

	bs, _ := json.MarshalIndent(meth.doc, "\t// ", "  ")
	pn("// %s\n", string(bs))
	pn("}")
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