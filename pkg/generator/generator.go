package generator

import (
	"github.com/go-openapi/spec"
	"bytes"
	"fmt"
	"go/format"
)

type genContext struct {
	spec        *spec.Swagger
	namespace string
	buf *bytes.Buffer
}

func (ctx *genContext)P(format string, a ...interface{}){
	ctx.buf.WriteString(fmt.Sprintf(format,a...))
}

func (ctx *genContext)Pn(format string, a ...interface{}) {
	ctx.buf.WriteString(fmt.Sprintf(format, a...) + "\n")
}

func (ctx *genContext)genServer()(code string,err error) {
	ctx.Pn("package %s", ctx.namespace)
	ctx.Pn("")
	ctx.Pn("import(")
	ctx.Pn(")")

	formattedCode, err := format.Source(ctx.buf.Bytes())
	if err != nil {
		return "", err
	}

	return string(formattedCode), nil
}

func GenerateServer(spec *spec.Swagger,namespace string) (code string,err error) {
	ctx := &genContext{
		spec:spec,
		namespace:namespace,
		buf:bytes.NewBufferString(""),
	}

	return ctx.genServer()
}