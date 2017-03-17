package generate

import (
	"bytes"
	"fmt"
	"log"
)

type Argument struct {
	method           *Method
	apiname, apitype string
	goname, gotype   string
	location         string // "path", "query", "body"
	desc             string

	required         bool
}

func (a *Argument) String() string {
	return a.goname + " " + a.gotype
}

func (a *Argument) ExprAsString(prefix string) string {
	switch a.gotype {
	case "[]string":
		log.Printf("TODO(bradfitz): only including the first parameter in path query.")
		return prefix + a.goname + `[0]`
	case "string":
		return prefix + a.goname
	case "integer", "int64":
		return "strconv.FormatInt(" + prefix + a.goname + ", 10)"
	case "uint64":
		return "strconv.FormatUint(" + prefix + a.goname + ", 10)"
	}
	log.Panicf("unknown type: apitype=%q, gotype=%q", a.apitype, a.gotype)
	return ""
}

type Arguments struct {
	l      []*Argument
	m      map[string]*Argument
	method *Method
}

func (args *Arguments) ForLocation(loc string) []*Argument {
	matches := make([]*Argument, 0)
	for _, arg := range args.l {
		if arg.location == loc {
			matches = append(matches, arg)
		}
	}
	return matches
}

func (args *Arguments) BodyArg() *Argument {
	for _, arg := range args.l {
		if arg.location == "body" {
			return arg
		}
	}
	return nil
}

func (args *Arguments) AddArg(arg *Argument) {
	n := 1
	oname := arg.goname
	for {
		_, present := args.m[arg.goname]
		if !present {
			args.m[arg.goname] = arg
			args.l = append(args.l, arg)
			return
		}
		n++
		arg.goname = fmt.Sprintf("%s%d", oname, n)
	}
}

func (a *Arguments) String() string {
	var buf bytes.Buffer
	for i, arg := range a.l {
		if i != 0 {
			buf.Write([]byte(", "))
		}
		buf.Write([]byte(arg.String()))
	}
	return buf.String()
}
