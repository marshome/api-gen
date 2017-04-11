package genlib

import (
	"net/http"
	"github.com/marshome/3p-mux"
	"encoding/json"
	"net/url"
)

type Context struct {
	HttpResponseWriter http.ResponseWriter
	HttpRequest        *http.Request

	MethodOptions      *MethodOptions
	PathParamMap       map[string]string
	CommonOptions      *CommonOptions

	ServiceResponse    interface{}
	ServiceError       error
}

type MethodOptions struct {
	Scopes []string
}

type CommonOptions struct {

}

func ParseCommonOptions(values url.Values) (opts *CommonOptions, err error) {
	opts = &CommonOptions{}

	return opts, nil
}

type Router interface {
	Handle(method string, path string, handlerFunc func(ctx *Context), methodOptions *MethodOptions)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type muxRouter struct {
	router        *mux.Router

	preprocessor  func(ctx *Context)
	postProcessor func(ctx*Context)
}

func (r *muxRouter)Handle(httpMethod string, httpPath string, handlerFunc func(ctx *Context), methodOptions *MethodOptions) {
	r.router.HandleFunc(httpPath, func(w http.ResponseWriter, req *http.Request) {
		ctx := &Context{
			HttpResponseWriter:w,
			HttpRequest:req,
			MethodOptions:methodOptions,
		}

		ctx.PathParamMap = mux.Vars(req)

		commonOptions, err := ParseCommonOptions(req.URL.Query())
		if err != nil {
			return
		}
		ctx.CommonOptions = commonOptions

		if r.preprocessor != nil {
			r.preprocessor(ctx)
		}

		if handlerFunc != nil {
			handlerFunc(ctx)
		}

		if r.postProcessor != nil {
			r.postProcessor(ctx)
		}

		if ctx.ServiceError != nil {
			w.Write([]byte(ctx.ServiceError.Error()))
			return
		}

		if ctx.ServiceResponse != nil {
			data, err := json.MarshalIndent(ctx.ServiceResponse, "", "    ")
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			w.Write(data)
		}else{
			w.Write([]byte("ok"))
		}
	}).Methods(httpMethod)
}

func (r *muxRouter)ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func NewRouter() Router {
	r:= &muxRouter{
		router:mux.NewRouter(),
	}

	return r
}
