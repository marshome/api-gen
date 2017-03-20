package marsapi

import (
	"net/http"
	"github.com/gorilla/mux"
	"net/url"
)

type ServerResponse struct {
	HttpStatusCode int
	HttpHeader     http.Header
}

type Context struct {
	HttpResponseWriter http.ResponseWriter
	HttpRequest        *http.Request
	ServiceRequest     interface{}
	ServiceResponse    interface{}
	Error              error
	PathParamMap         map[string]string
}

type Router interface {
	Handle(method string, path string, handlerFunc func(ctx *Context))
}

type muxRouter struct {
	router        mux.Router

	preprocessor  func(ctx *Context)
	postProcessor func(ctx*Context)
}

func (r *muxRouter)Handle(method string, path string,handlerFunc func(ctx *Context)) {
	r.router.HandleFunc(path, func(httpResponseWriter http.ResponseWriter, httpRequest *http.Request) {
		ctx := &Context{
			HttpResponseWriter:httpResponseWriter,
			HttpRequest:httpRequest,
			PathParamMap:mux.Vars(httpRequest),
		}

		if r.preprocessor != nil {
			r.preprocessor(ctx)
		}

		if handlerFunc != nil {
			handlerFunc(ctx)
		}

		if r.postProcessor != nil {
			r.postProcessor(ctx)
		}

		ctx.HttpRequest.URL.Query().Encode()
	})
}

func NewRouter() Router {
	return &muxRouter{}
}