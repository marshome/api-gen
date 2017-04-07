package googlespec

import (
	"github.com/marshome/apis/spec"
	"strings"
)

func Convert(doc *APIDocument)(s *spec.Document) {
	s = &spec.Document{}

	s.Kind = doc.Kind
	s.ETag = doc.ETag
	s.DiscoveryVersion = doc.DiscoveryVersion
	s.Id = doc.Id
	s.Name = doc.Name
	s.Version = doc.Version
	s.Title = doc.Title
	s.Description = doc.Description
	s.DocumentationLink = doc.DocumentationLink
	s.Labels = doc.Labels
	s.Protocol = doc.Protocol
	s.RootUrl = doc.RootUrl
	s.Features = doc.Features

	s.Auth = &spec.Auth{}
	if doc.Auth != nil {
		if doc.Auth.OAuth2 != nil &&doc.Auth.OAuth2.Scopes != nil {
			s.Auth.OAuth2Scopes = make([]string, 0)
			for i, _ := range doc.Auth.OAuth2.Scopes {
				s.Auth.OAuth2Scopes = append(s.Auth.OAuth2Scopes, i)
			}
		}
	}

	if doc.Parameters != nil {
		s.Parameters = make([]*spec.Object, 0)
		for name, obj := range doc.Parameters {
			s.Parameters = append(s.Parameters, convertObject(name, obj))
		}
	}

	if doc.Schemas != nil {
		s.Schemas = make([]*spec.Object, 0)
		for name, obj := range doc.Schemas {
			s.Schemas = append(s.Schemas, convertObject(name, obj))
		}
	}

	if doc.Resources != nil {
		s.Resources = make([]*spec.Resource, 0)
		for rName, r := range doc.Resources {
			s.Resources=append(s.Resources,convertResource(rName,r))
		}
	}

	return s
}

func convertType(typ string,format string)string {
	if typ == "string" {
		if format == "" {
			return spec.TYPE_STIRNG
		} else if format == "int64" {
			return spec.TYPE_INT64
		} else if format == "date-time" {
			return spec.TYPE_DATETIME
		}else if format=="byte"{
			return spec.TYPE_BYTE
		}else if format=="google-datetime"{
			return spec.TYPE_DATETIME
		}else if format=="uint64"{
			return spec.TYPE_UINT64
		}else if format=="google-duration"{
			return spec.TYPE_STIRNG
		}else if format=="google-fieldmask"{
			return spec.TYPE_STIRNG
		}else if format=="date"{
			return spec.TYPE_DATE
		}
	} else if typ == "boolean" {
		if format == "" {
			return spec.TYPE_BOOL
		}
	} else if typ == "integer" {
		if format == "int32" {
			return spec.TYPE_INT32
		} else if format == "uint32" {
			return spec.TYPE_UINT32
		}
	} else if typ == "number" {
		if format == "double" {
			return spec.TYPE_FLOAT64
		}else if format=="float"{
			return spec.TYPE_FLOAT32
		}
	} else if typ == "any" {
		if format == "" {
			return spec.TYPE_ANY
		}
	}

	panic("convertType " + typ + " " + format)
}

func convertObject(name string,doc *APIObject)*spec.Object {
	o := &spec.Object{}
	o.Name = name
	o.Desc = doc.Description
	o.Required = doc.Required
	o.Default = doc.Default
	o.Pattern = doc.Pattern
	o.Min = doc.Minimum
	o.Max = doc.Maximum
	if doc.Enum != nil {
		o.Enum = make([]*spec.Enum, 0)
		for i, v := range doc.Enum {
			desc := ""
			if doc.EnumDescriptions != nil&&i < len(doc.EnumDescriptions) {
				desc = doc.EnumDescriptions[i]
			}
			o.Enum = append(o.Enum, &spec.Enum{Name:v, Desc:desc})
		}
	}

	if doc.Type == "array" {
		//array
		if doc.Items == nil {
			panic("array no items " + name)
		}
		o.Collection = spec.COLLECTION_ARRAY
		collectionType := convertObject("", doc.Items)
		if collectionType.Collection != "" {
			o.CollectionItem = collectionType
		} else {
			o.Type = collectionType.Type
			if o.Type == spec.TYPE_REF {
				o.RefType = collectionType.RefType
			} else if o.Type == spec.TYPE_OBJECT {
				o.Fields = collectionType.Fields
			}
		}
	} else if doc.AdditionalProperties != nil {
		//map
		o.Collection = spec.COLLECTION_MAP
		collectionType := convertObject("", doc.AdditionalProperties)
		if collectionType.Collection != "" {
			o.CollectionItem = collectionType
		} else {
			o.Type = collectionType.Type
			if o.Type == spec.TYPE_REF {
				o.RefType = collectionType.RefType
			} else if o.Type == spec.TYPE_OBJECT {
				o.Fields = collectionType.Fields
			}
		}
	} else if doc.Ref != "" {
		//ref
		o.Type = "ref"
		o.RefType = doc.Ref
	} else if doc.Type == "object" {
		//obj
		o.Type = "object"
		if doc.Properties != nil {
			o.Fields = make([]*spec.Object, 0)
			for fName, fSpec := range doc.Properties {
				if strings.HasPrefix(fName, "@") {
					continue
				}
				o.Fields = append(o.Fields, convertObject(fName, fSpec))
			}
		}
	} else {
		//simple
		o.Type = convertType(doc.Type, doc.Format)
	}

	return o
}

func convertResource(name string,doc *APIResource)*spec.Resource {
	r := &spec.Resource{}
	r.Name = name

	if doc.Methods != nil {
		r.Methods = make([]*spec.Method, 0)
		for mName, m := range doc.Methods {
			r.Methods = append(r.Methods, convertMethod(mName, m))
		}
	}

	if doc.Resources!=nil{
		r.Resources=make([]*spec.Resource,0)
		for subName,subR:=range doc.Resources{
			r.Resources=append(r.Resources,convertResource(subName,subR))
		}
	}

	return r
}

func convertMethod(name string,doc *APIMethod)*spec.Method {
	m := &spec.Method{}
	m.Name = name
	m.Desc = doc.Description
	m.Path = doc.Path
	m.HttpMethod = doc.HttpMethod
	m.Scopes = doc.Scopes

	m.PathParams = make([]*spec.Object, 0)
	m.RequiredQueryParams = make([]*spec.Object, 0)
	m.OptionalQueryParams = make([]*spec.Object, 0)
	if doc.ParameterOrder != nil {
		for _, v := range doc.ParameterOrder {
			p := doc.Parameters[v]
			if p.Location == "path" {
				m.PathParams = append(m.PathParams, convertObject(v, p))
			} else if p.Location == "query" {
				m.RequiredQueryParams = append(m.RequiredQueryParams, convertObject(v, p))
			} else {
				panic("location " + p.Location)
			}
		}
	}

	for i, v := range doc.Parameters {
		if !v.Required {
			m.OptionalQueryParams = append(m.OptionalQueryParams, convertObject(i, v))
		}
	}

	if doc.Request != nil {
		m.Request = doc.Request.Ref
	}

	if doc.Response != nil {
		m.Response = doc.Response.Ref
	}

	return m
}
