package goolge2swagger

import (
	"github.com/go-openapi/spec"
	"github.com/marshome/i-api/spec/googlespec"
)

func ConvertParameter(g *googlespec.APIObject) (s *spec.Parameter, err error) {
	return &spec.Parameter{}, nil
}

func ConvertSchema(g *googlespec.APIObject) (s *spec.Schema, err error) {
	s = &spec.Schema{}

	s.ID = g.Id

	if g.Ref != "" {
		s.Ref, err = spec.NewRef(g.Ref)
		if err != nil {
			return nil, err
		}
	}

	s.Description = g.Description

	s.AddType(g.Type, g.Format)

	if g.Properties != nil {
		s.Properties = make(map[string]spec.Schema)
		for k, v := range g.Properties {
			p, err := ConvertSchema(v)
			if err != nil {
				return nil, err
			}
			s.Properties[k] = *p
		}
	}

	if g.AdditionalProperties != nil {
		schema, err := ConvertSchema(g.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		s.AdditionalProperties = &spec.SchemaOrBool{
			Schema: schema,
		}
	}

	if g.Items != nil {
		schema, err := ConvertSchema(g.Items)
		if err != nil {
			return nil, err
		}
		s.Items = &spec.SchemaOrArray{
			Schema: schema,
		}
	}

	return s, nil
}

func ConvertResourceRecursive(name string, g *googlespec.APIResource, paths *spec.Paths) (err error) {
	if g.Methods != nil {
		p := spec.PathItem{}
		paths.Paths[name] = p
		for _, m := range g.Methods {
			if m.HttpMethod == "GET" {
				p.Get = &spec.Operation{}
			}
		}
	}

	for subRName, subR := range g.Resources {
		ConvertResourceRecursive(subRName, subR, paths)
	}

	return nil
}

func Convert(g *googlespec.APIDocument) (s *spec.Swagger, err error) {
	s = &spec.Swagger{}

	s.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Description: g.Description,
			Title:       g.Title,
			Contact: &spec.ContactInfo{
				Name: g.OwnerName,
				URL:  g.OwnerDomain,
			},
			Version: g.Version,
		},
	}
	s.Host = g.RootUrl
	s.BasePath = g.ServicePath

	if g.Parameters != nil {
		s.Parameters = make(map[string]spec.Parameter, 0)
		for k, v := range g.Parameters {
			param, err := ConvertParameter(v)
			if err != nil {
				return nil, err
			}
			s.Parameters[k] = *param
		}
	}

	if g.Schemas != nil {
		s.Definitions = make(spec.Definitions, 0)
		for k, v := range g.Schemas {
			schema, err := ConvertSchema(v)
			if err != nil {
				return nil, err
			}
			s.Definitions[k] = *schema
		}
	}

	s.Paths = &spec.Paths{}
	if g.Resources != nil {
		s.Paths.Paths = make(map[string]spec.PathItem, 0)
		for k, v := range g.Resources {
			ConvertResourceRecursive(k, v, s.Paths)
		}
	}

	return s, nil
}
