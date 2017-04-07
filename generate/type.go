package generate

import (
	"fmt"
	"strings"

	"github.com/marshome/apis/googlespec"
)

type Type struct {
	c        *Context
	m        *googlespec.APIObject
	_apiName string
}

func (t *Type) GoType() string {

	if t, ok := t.c.SimpleTypeConvert(t.m.Type, t.m.Format); ok {
		return t
	}


	if at := t.ArrayType(); at != nil {

		if at.m.Type == "string" {
			switch at.m.Format {
			case "int64":
				panic(11)
				return "marsapi.Int64s"
			case "uint64":
				panic(22)
				return "marsapi.Uint64s"
			case "int32":
				panic(33)
				return "marsapi.Int32s"
			case "uint32":
				panic(44)
				return "marsapi.Uint32s"
			case "float64":
				panic(55)
				return "marsapi.Float64s"
			default:
				return "[]" + at.GoType()
			}
		}
		return "[]" + at.GoType()
	}
	if ref, ok := t.Reference(); ok {
		s := t.c.Schemas[ref]
		if s == nil {
			panic(fmt.Sprintf("in Type.AsGo(), failed to find referenced type %q for %s",
				ref, t.c.PrettyJSON(t.m)))
		}

		return "*" + s.GoName
	}

	if mt := t.MapType(); mt != nil {
		return mt.GoType()
	}

	if t.IsStruct() {
		if t._apiName != "" {
			s := t.c.Schemas[t._apiName]
			if s == nil {
				panic(fmt.Sprintf("in Type.AsGo, _apiName of %q didn't point to a valid schema; json: %s",
					t._apiName, t.c.PrettyJSON(t.m)))
			}

			return "*" + s.GoName
		}
		fmt.Println("Type.AsGo", t)
		panic("in Type.AsGo, no _apiName found for struct type " + t.c.PrettyJSON(t.m))
	}

	panic("unhandled Type.AsGo for " + t.c.PrettyJSON(t.m))
}

func (t *Type) MapType() (typ *Type) {
	if t.m.AdditionalProperties == nil {
		return nil
	}

	return &Type{c:t.c, m:t.m.AdditionalProperties, _apiName:t._apiName}
}

func (t *Type) ArrayType()  *Type {
	if t.m.Type != "array" {
		return nil
	}

	items := t.m.Items
	if items == nil {
		panic(fmt.Sprintf("can't handle array type missing its 'items' key. map is %#v", t.m))
	}

	return &Type{c: t.c, m: items, _apiName:t._apiName}
}

func (t *Type) IsIntAsString() bool {
	return t.m.Type == "string" && strings.Contains(t.m.Format, "int")
}

func (t *Type) IsSimple() bool {
	_, ok := t.c.SimpleTypeConvert(t.m.Type , t.m.Format)
	return ok
}

func (t *Type) IsStruct() bool {
	return t.m.Type == "object"&&t.m.AdditionalProperties == nil
}

func (t *Type) Reference() (apiName string, ok bool) {
	apiName = t.m.Ref
	ok = apiName != ""
	return
}

func (t *Type) IsReference() bool {
	return t.m.Ref != ""
}