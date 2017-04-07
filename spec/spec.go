package spec

const(
	TYPE_STIRNG="string"
	TYPE_BOOL="bool"
	TYPE_BYTE="byte"
	TYPE_INT32="int32"
	TYPE_UINT32="uint32"
	TYPE_INT64="int64"
	TYPE_UINT64="uint64"
	TYPE_FLOAT32="float32"
	TYPE_FLOAT64="float64"
	TYPE_DATE="date"
	TYPE_DATETIME="datetime"
	TYPE_ANY="any"
	TYPE_REF="ref"
	TYPE_OBJECT="object"

	COLLECTION_NONE=""
	COLLECTION_ARRAY="array"
	COLLECTION_MAP="map"
)

type Directory struct {
	Kind             string              `json:"kind"`             //The fixed string discovery#directoryList
	DiscoveryVersion string              `json:"discoveryVersion"` //Indicate the version of the Discovery API used to generate this doc.
	Items            []*DirectoryItem `json:"items"`               //The individual directory entries. One entry per API/version pair.
}

type DirectoryItem struct {
	Kind              string   `json:"kind"`              //The kind for this response.
	Id                string   `json:"id"`                //The ID of this API.
	Name              string   `json:"name"`              ///The name of the API.
	Version           string   `json:"version"`           //The version of the API.
	Title             string   `json:"title"`             //The title of this API.
	Description       string   `json:"description"`       //The description of this API.
	DiscoveryRestUrl  string   `json:"discoveryRestUrl"`  //The url for the discovery REST document.
	DiscoveryLink     string   `json:"discoveryLink"`     //A link to the discovery document.
	DocumentationLink string   `json:"documentationLink"` //A link to human readable documentation for the API.
	Labels            []string `json:"labels"`            //Labels for the status of this API, such as limited_availability or deprecated.
	Preferred         bool     `json:"preferred"`         //true if this version is the preferred version to use.
}

type Document struct {
	Kind              string   `json:"kind"`                              // The kind for this response. The fixed string discovery#restDescription.
	ETag              string   `json:"etag"`
	DiscoveryVersion  string   `json:"discoveryVersion"`                  // Indicate the version of the Discovery API used to generate this doc.
	Id                string   `json:"id"`                                // The ID of the Discovery document for the API. For example, urlshortener:v1.
	Name              string   `json:"name"`                              // The name of the API. For example, urlshortener.
	Version           string   `json:"version"`                           // The version of the API. For example, v1.
	Title             string   `json:"title"`                             // The title of the API. For example, "Google Url Shortener API".
	Description       string   `json:"description"`                       // The description of this API.
	DocumentationLink string   `json:"documentationLink"`                 // A link to human-readable documentation for the API.
	Labels            []string `json:"labels,omitempty"`                  // Labels for the status of this API. Valid values include limited_availability or deprecated.
	Protocol          string   `json:"protocol"`                          // The protocol described by the document. For example, REST.
	RootUrl           string   `json:"rootUrl"`                           // The root url under which all API services live.

	Features          []string                `json:"features,omitempty"` // A list of supported features for this API.
	Auth              *Auth                `json:"auth,omitempty"`        // Authentication information.
	Parameters        []*Object   `json:"parameters,omitempty"`           // Common parameters that apply across all apis.
	Schemas           []*Object   `json:"schemas,omitempty"`              // The schemas for this API.
	Resources         []*Resource `json:"resources,omitempty"`            // object	The resources in this API.
}

type Auth struct {
	OAuth2Scopes []string `json:"oauth2,omitempty"`
}

type Enum struct {
	Name string `json:"name,omitempty"`
	Desc string `json:"desc,omitempty"`
}

type Object struct {
	Name           string                        `json:"name,omitempty"`
	Desc           string                        `json:"desc,omitempty"`
	Required       bool `json:"required,omitempty"`
	Collection     string `json:"collection,omitempty"`
	Type           string                `json:"type,omitempty"`
	Default        string        `json:"default,omitempty"`
	Pattern        string        `json:"pattern,omitempty"`
	Min            string        `json:"min,omitempty"`
	Max            string        `json:"max,omitempty"`
	Enum           []*Enum        `json:"enum,omitempty"`
	RefType        string `json:"refType,omitempty"`
	CollectionItem *Object       `json:"collectionItem,omitempty"` //只有嵌套的集合类才会设置
	Fields         []*Object      `json:"fields,omitempty"`
}

type Method struct {
	Name                string        `json:"name,omitempty"`
	Desc                string        `json:"desc,omitempty"`
	Path                string        `json:"path"`
	HttpMethod          string        `json:"httpMethod"`
	Scopes              []string      `json:"scopes,omitempty"`
	PathParams          []*Object     `json:"pathParams,omitempty"`
	RequiredQueryParams []*Object     `json:"requiredQueryParams,omitempty"`
	OptionalQueryParams []*Object     `json:"optinalQueryParams,omitempty"`
	Request             string        `json:"request,omitempty"`
	Response            string        `json:"response,omitempty"`
}

type Resource struct {
	Name      string                `json:"name,omitempty"`
	Desc      string  `json:"desc,omitempty"`
	Methods   []*Method    `json:"methods,omitempty"`
	Resources []*Resource  `json:"resources,omitempty"`
}
