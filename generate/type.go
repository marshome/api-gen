package generate

import (
	"fmt"
	"log"
	"strings"

	"github.com/marshome/apis/spec"
)

type Type struct {
	c        *Context
	m        *spec.APIObject // JSON map containing key "type" and maybe "items", "properties"
	_apiName string
}

func (t *Type) ApiType() string {
	// Note: returns "" on reference types
	return t.m.Type
}

func (t *Type) ApiTypeFormat() string {
	return t.m.Format
}

func (t *Type) IsIntAsString() bool {
	return t.ApiType() == "string" && strings.Contains(t.ApiTypeFormat(), "int")
}

func (t *Type) AsSimpleGoType() (goType string, ok bool) {
	return t.c.SimpleTypeConvert(t.ApiType(), t.ApiTypeFormat())
}

func (t *Type) String() string {
	return fmt.Sprintf("[type=%q, map=%s]", t.ApiType(), t.c.PrettyJSON(t.m))
}

func (t *Type) AsGo() string {
	if t, ok := t.AsSimpleGoType(); ok {
		return t
	}
	if at, ok := t.ArrayType(); ok {
		if at.ApiType() == "string" {
			switch at.ApiTypeFormat() {
			case "int64":
				return "googleapi.Int64s"
			case "uint64":
				return "googleapi.Uint64s"
			case "int32":
				return "googleapi.Int32s"
			case "uint32":
				return "googleapi.Uint32s"
			case "float64":
				return "googleapi.Float64s"
			default:
				return "[]" + at.AsGo()
			}
		}
		return "[]" + at.AsGo()
	}
	if ref, ok := t.Reference(); ok {
		s := t.c.Schemas[ref]
		if s == nil {
			panic(fmt.Sprintf("in Type.AsGo(), failed to find referenced type %q for %s",
				ref, t.c.PrettyJSON(t.m)))
		}
		return s.Type().AsGo()
	}
	if typ, ok := t.MapType(); ok {
		return typ
	}
	isAny := t.IsAny()
	if t.IsStruct() || isAny {
		if t._apiName != "" {
			s := t.c.Schemas[t._apiName]
			if s == nil {
				panic(fmt.Sprintf("in Type.AsGo, _apiName of %q didn't point to a valid schema; json: %s",
					t._apiName, t.c.PrettyJSON(t.m)))
			}
			if isAny {
				return s.GoName() // interface type; no pointer.
			}
			//if v := jobj(s.m, "variant"); v != nil {//todo
			//	return s.GoName()
			//}
			return "*" + s.GoName()
		}
		panic("in Type.AsGo, no _apiName found for struct type " + t.c.PrettyJSON(t.m))
	}
	panic("unhandled Type.AsGo for " + t.c.PrettyJSON(t.m))
}

func (t *Type) IsSimple() bool {
	_, ok := t.c.SimpleTypeConvert(t.ApiType(), t.ApiTypeFormat())
	return ok
}

func (t *Type) IsStruct() bool {
	return t.ApiType() == "object"
}

func (t *Type) IsAny() bool {
	if t.ApiType() == "object" {
		props := t.m.AdditionalProperties
		if props != nil && props.Type == "any" {
			return true
		}
	}
	return false
}

func (t *Type) Reference() (apiName string, ok bool) {
	apiName = t.m.Ref
	ok = apiName != ""
	return
}

func (t *Type) IsMap() bool {
	_, ok := t.MapType()
	return ok
}

// MapType checks if the current node is a map and if true, it returns the Go type for the map, such as map[string]string.
func (t *Type) MapType() (typ string, ok bool) {
	props := t.m.AdditionalProperties
	if props == nil {
		return "", false
	}
	s := props.Type
	if s == "any" {
		return "", false
	}
	if s == "string" {
		return "map[string]string", true
	}
	if s != "array" {
		if s == "" { // Check for reference
			s = props.Ref
			if s != "" {
				return "map[string]" + s, true
			}
		}
		if s == "any" {
			return "map[string]interface{}", true
		}
		log.Printf("Warning: found map to type %q which is not implemented yet.", s)
		return "", false
	}
	items := props.Items
	if items == nil {
		return "", false
	}
	s = items.Type
	if s != "string" {
		if s == "" { // Check for reference
			s = items.Ref
			if s != "" {
				return "map[string][]" + s, true
			}
		}
		log.Printf("Warning: found map of arrays of type %q which is not implemented yet.", s)
		return "", false
	}
	return "map[string][]string", true
}

func (t *Type) IsReference() bool {
	return t.m.Ref != ""
}

func (t *Type) ReferenceSchema() (s *Schema, ok bool) {
	apiName, ok := t.Reference()
	if !ok {
		return
	}

	s = t.c.Schemas[apiName]
	if s == nil {
		t.c.Panicf("failed to find t.api.schemas[%q] while resolving reference",
			apiName)
	}
	return s, true
}

func (t *Type) ArrayType() (elementType *Type, ok bool) {
	if t.ApiType() != "array" {
		return
	}
	items := t.m.Items
	if items == nil {
		t.c.Panicf("can't handle array type missing its 'items' key. map is %#v", t.m)
	}
	return &Type{c: t.c, m: items}, true
}
