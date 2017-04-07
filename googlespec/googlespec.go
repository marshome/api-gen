package googlespec

type APIDirectory struct {
	Kind             string              `json:"kind"`             //The fixed string discovery#directoryList
	DiscoveryVersion string              `json:"discoveryVersion"` //Indicate the version of the Discovery API used to generate this doc.
	Items            []*APIDirectoryItem `json:"items"`            //The individual directory entries. One entry per API/version pair.
}

type APIDirectoryItem struct {
	Kind              string   `json:"kind"`              //The kind for this response.
	Id                string   `json:"id"`                //The ID of this API.
	Name              string   `json:"name"`              ///The name of the API.
	Version           string   `json:"version"`           //The version of the API.
	Title             string   `json:"title"`             //The title of this API.
	Description       string   `json:"description"`       //The description of this API.
	DiscoveryRestUrl  string   `json:"discoveryRestUrl"`  //The url for the discovery REST document.
	DiscoveryLink     string   `json:"discoveryLink"`     //A link to the discovery document.
	Icons             *APIIcon `json:"icons"`             //Links to 16x16 and 32x32 icons representing the API.
	DocumentationLink string   `json:"documentationLink"` //A link to human readable documentation for the API.
	Labels            []string `json:"labels"`            //Labels for the status of this API, such as limited_availability or deprecated.
	Preferred         bool     `json:"preferred"`         //true if this version is the preferred version to use.
}

type APIDocument struct {
	Kind              string   `json:"kind"` //The kind for this response. The fixed string discovery#restDescription.
	ETag              string   `json:"etag"`
	DiscoveryVersion  string   `json:"discoveryVersion"` //Indicate the version of the Discovery API used to generate this doc.
	Id                string   `json:"id"`               //The ID of the Discovery document for the API. For example, urlshortener:v1.
	Name              string   `json:"name"`             //The name of the API. For example, urlshortener.
	CanonicalName     string   `json:"canonicalName"`    //人类可读的名字
	Version           string   `json:"version"`          //The version of the API. For example, v1.
	Revision          string   `json:"revision"`         //The revision of the API.
	Title             string   `json:"title"`            //The title of the API. For example, "Google Url Shortener API".
	Description       string   `json:"description"`      //The description of this API.
	OwnerDomain       string   `json:"ownerDomain"`
	OwnerName         string   `json:"ownerName"`
	Icons             *APIIcon `json:"icons"`             //Links to 16x16 and 32x32 icons representing the API.
	DocumentationLink string   `json:"documentationLink"` //A link to human-readable documentation for the API.
	Labels            []string `json:"labels,omitempty"`  //Labels for the status of this API. Valid values include limited_availability or deprecated.
	Protocol          string   `json:"protocol"`          //The protocol described by the document. For example, REST.
	RootUrl           string   `json:"rootUrl"`           //The root url under which all API services live.
	ServicePath       string   `json:"servicePath"`       //The base path for all REST requests.
	BatchPath         string   `json:"batchPath"`         //The path for REST batch requests.

	Parameters map[string]*APIObject   `json:"parameters,omitempty"` //Common parameters that apply across all apis.
	Auth       *APIAuth                `json:"auth,omitempty"`       //Authentication information.
	Features   []string                `json:"features,omitempty"`   //A list of supported features for this API.
	Schemas    map[string]*APIObject   `json:"schemas,omitempty"`    //The schemas for this API.
	Methods    map[string]*APIMethod   `json:"methods,omitempty"`    //API-level methods for this API.
	Resources  map[string]*APIResource `json:"resources,omitempty"`  //object	The resources in this API.
}

type APIIcon struct {
	X16 string `json:"x16,omitempty"` //The URL of the 16x16 icon.
	X32 string `json:"x32,omitempty"` //The URL of the 32x32 icon.
}

type APIAnnotation struct {
	Required []string `json:"required"` //A list of methods that require this property on requests.
}

type APIObject struct {
	Id                   string                `json:"id,omitempty"`                   //Unique identifier for this schema.
	Type                 string                `json:"type,omitempty"`                 //The value type for this schema. A list of values can be found at the "type" section in the JSON Schema.
	Ref                  string                `json:"$ref,omitempty"`                 //A reference to another schema. The value of this property is the ID of another schema.
	Description          string                `json:"description,omitempty"`          //A description of this object.
	Default              string                `json:"default,omitempty"`              //The default value of this property (if one exists).
	Required             bool                  `json:"required,omitempty"`             //Whether the parameter is required.
	Format               string                `json:"format,omitempty"`               //An additional regular expression or key that helps constrain the value. For more details see the Type and Format Summary.
	Pattern              string                `json:"pattern,omitempty"`              //The regular expression this parameter must conform to.
	Minimum              string                `json:"minimum,omitempty"`              //The minimum value of this parameter.
	Maximum              string                `json:"maximum,omitempty"`              //The maximum value of this parameter.
	Enum                 []string              `json:"enum,omitempty"`                 //Values this parameter may take (if it is an enum).
	EnumDescriptions     []string              `json:"enumDescriptions,omitempty"`     //The descriptions for the enums. Each position maps to the corresponding value in the enum array.
	Repeated             bool                  `json:"repeated,omitempty"`             //Whether this parameter may appear multiple times.
	Location             string                `json:"location,omitempty"`             //Whether this parameter goes in the query or the path for REST requests.
	Properties           map[string]*APIObject `json:"properties,omitempty"`           //If this is a schema for an object, list the schema for each property of this object.	The value is itself a JSON Schema object describing this property.
	AdditionalProperties *APIObject            `json:"additionalProperties,omitempty"` //If this is a schema for an object, this property is the schema for any additional properties with dynamic keys on this object.
	Items                *APIObject            `json:"items,omitempty"`                //If this is a schema for an array, this property is the schema for each element in the array.
	Annotations          *APIAnnotation        `json:"annotations,omitempty"`          //Additional information about this property.
}

type APIAuth struct {
	OAuth2 *APIOAuth2 `json:"oauth2,omitempty"`
}

type APIOAuth2 struct {
	Scopes map[string]*APIOAuth2Scope `json:"scopes"` //Available OAuth 2.0 scopes.
}

type APIOAuth2Scope struct {
	Description string `json:"description"` //Description of scope.
}

type APIMethod struct {
	Id                    string                `json:"id"`                              //A unique ID for this method. This property can be used to match methods between different versions of Discovery.
	Path                  string                `json:"path"`                            //The URI path of this REST method. Should be used in conjunction with the servicePath property at the API-level.
	HttpMethod            string                `json:"httpMethod"`                      //HTTP method used by this method.
	Description           string                `json:"description,omitempty"`           //Description of this method.
	Parameters            map[string]*APIObject `json:"parameters,omitempty"`            //Details for all parameters in this method.
	ParameterOrder        []string              `json:"parameterOrder,omitempty"`        //Ordered list of required parameters. This serves as a hint to clients on how to structure their method signatures. The array is ordered such that the most significant parameter appears first.
	Request               *APIMethodReqest      `json:"request,omitempty"`               //The schema for the request.
	Response              *APIMethodResponse    `json:"response,omitempty"`              //The schema for the response.
	Scopes                []string              `json:"scopes,omitempty"`                //OAuth 2.0 scopes applicable to this method.
	SupportsMediaDownload bool                  `json:"supportsMediaDownload,omitempty"` //Whether this method supports media downloads.
	SupportsMediaUpload   bool                  `json:"supportsMediaUpload,omitempty"`   //Whether this method supports media uploads.
	SupportsSubscription  bool                  `json:"supportsSubscription,omitempty"`  //Whether this method supports subscriptions.
	MediaUpload           *APIMethodMediaUpload `json:"mediaUpload,omitempty"`           //Media upload parameters.
}

type APIMethodReqest struct {
	Ref string `json:"$ref"` //Schema ID for the request schema.
}

type APIMethodResponse struct {
	Ref string `json:"$ref"` //Schema ID for the response schema.
}

type APIMethodMediaUpload struct {
	Accept    []string                      `json:"accept,omitempty"`    //MIME Media Ranges for acceptable media uploads to this method.
	MaxSize   string                        `json:"maxSize,omitempty"`   //Maximum size of a media upload, such as "1MB", "2GB" or "3TB".
	Protocols *APIMethodMediaUploadProtocal `json:"protocols,omitempty"` //Supported upload protocols.
}

type APIMethodMediaUploadProtocal struct {
	Simple    *APIMethodMediaUploadProtocalSimple    `json:"simple,omitempty"`    //Supports uploading as a single HTTP request.
	Resumable *APIMethodMediaUploadProtocalResumable `json:"resumable,omitempty"` //Supports the Resumable Media Upload protocol.
}

type APIMethodMediaUploadProtocalSimple struct {
	Multipart bool   `json:"multipart,omitempty"` //True if this endpoint supports upload multipart media.
	Path      string `json:"path,omitempty"`      //The URI path to be used for upload. Should be used in conjunction with the rootURL property at the api-level.
}

type APIMethodMediaUploadProtocalResumable struct {
	Multipart bool   `json:"multipart,omitempty"` //true if this endpoint supports uploading multipart media.
	Path      string `json:"path,omitempty"`      //The URI path to be used for upload. Should be used in conjunction with the rootURL property at the API-level.
}

type APIResource struct {
	Methods   map[string]*APIMethod   `json:"methods,omitempty"`   //Methods on this resource.
	Resources map[string]*APIResource `json:"resources,omitempty"` //Sub-resources on this resource.
}
