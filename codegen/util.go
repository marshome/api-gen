package codegen

import (
	"strings"
	"github.com/marshome/apis/spec"
	"regexp"
	"fmt"
	"bytes"
	"unicode"
)

var urlRE = regexp.MustCompile(`^http\S+$`)

func Comment(pfx, c string) string {
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
			fmt.Fprintf(&buf, "%s// %s", pfx, r.Replace(line))
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

func GoName(name string, initialCap bool) string {
	var buf bytes.Buffer
	preserve_ := false
	for i, c := range name {
		if c == '_' {
			if preserve_ || strings.HasPrefix(name[i:], "__") {
				preserve_ = true
			} else {
				initialCap = true
				continue
			}
		} else {
			preserve_ = false
		}
		if c == '-' || c == '.' || c == '$' || c == '/' {
			initialCap = true
			continue
		}
		if initialCap {
			c = unicode.ToUpper(c)
			initialCap = false
		}
		buf.WriteByte(byte(c))
	}

	goName := buf.String()
	_, has := go_tokens[goName]
	if has {
		return goName + "_"
	}

	return goName
}

func GoType(obj * spec.Object,parentName string,topCollectionItem bool)string {
	typ := ""
	switch obj.Type {
	case "":
		if obj.Collection == "" || obj.CollectionItem == nil {
			panic("unknown type " + obj.Name + " " + obj.Type)
		}
		typ = GoType(obj.CollectionItem, parentName, topCollectionItem)
	case spec.TYPE_STRING:typ = "string"
	case spec.TYPE_BOOL:typ = "bool"
	case spec.TYPE_BYTE:typ = "byte"
	case spec.TYPE_INT32:typ = "int32"
	case spec.TYPE_UINT32:typ = "uint32"
	case spec.TYPE_INT64:typ = "int64"
	case spec.TYPE_UINT64:typ = "uint64"
	case spec.TYPE_FLOAT32:typ = "float32"
	case spec.TYPE_FLOAT64:typ = "float64"
	case spec.TYPE_DATE:typ = "time.Time"
	case spec.TYPE_DATETIME:typ = "time.Time"
	case spec.TYPE_ANY:typ = "interface{}"
	case spec.TYPE_REF:typ = "*" + GoName(obj.RefType, true)
	case spec.TYPE_OBJECT:
		typ = "*" + GoName(parentName + "." + obj.Name, true)
		if topCollectionItem {
			typ += "Item"
		}
	default:
		panic("unknown type " + obj.Name + " " + obj.Type)
	}

	if obj.Collection == spec.COLLECTION_ARRAY {
		return "[]" + typ
	} else if obj.Collection == spec.COLLECTION_MAP {
		return "map[string]" + typ
	} else {
		return typ
	}
}