package marsapi

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strings"
	"bytes"
	"context"
)

// CloseBody is used to close res.Body.
// Prior to calling Close, it also tries to Read a small amount to see an EOF.
// Not seeing an EOF can prevent HTTP Transports from reusing connections.
func CloseBody(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	// Justification for 3 byte reads: two for up to "\r\n" after
	// a JSON/XML document, and then 1 to see EOF if we haven't yet.
	// TODO(bradfitz): detect Go 1.3+ and skip these reads.
	// See https://codereview.appspot.com/58240043
	// and https://codereview.appspot.com/49570044
	buf := make([]byte, 1)
	for i := 0; i < 3; i++ {
		_, err := res.Body.Read(buf)
		if err != nil {
			break
		}
	}
	res.Body.Close()

}

// Error contains an error response from the server.
type Error struct {
	// Code is the HTTP response status code and will always be populated.
	Code int `json:"code"`
	// Message is the server response message and is only populated when
	// explicitly referenced by the JSON server response.
	Message string `json:"message"`
	// Body is the raw response returned by the server.
	// It is often but not always JSON, depending on how the request fails.
	Body string
	// Header contains the response header fields from the server.
	Header http.Header

	Errors []ErrorItem
}

// ErrorItem is a detailed error code & message from the Google API frontend.
type ErrorItem struct {
	// Reason is the typed error code. For example: "some_example".
	Reason string `json:"reason"`
	// Message is the human-readable description of the error.
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if len(e.Errors) == 0 && e.Message == "" {
		return fmt.Sprintf("googleapi: got HTTP response code %d with body: %v", e.Code, e.Body)
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "googleapi: Error %d: ", e.Code)
	if e.Message != "" {
		fmt.Fprintf(&buf, "%s", e.Message)
	}
	if len(e.Errors) == 0 {
		return strings.TrimSpace(buf.String())
	}
	if len(e.Errors) == 1 && e.Errors[0].Message == e.Message {
		fmt.Fprintf(&buf, ", %s", e.Errors[0].Reason)
		return buf.String()
	}
	fmt.Fprintln(&buf, "\nMore details:")
	for _, v := range e.Errors {
		fmt.Fprintf(&buf, "Reason: %s, Message: %s\n", v.Reason, v.Message)
	}
	return buf.String()
}

type errorReply struct {
	Error *Error `json:"error"`
}

// CheckResponse returns an error (of type *Error) if the response
// status code is not 2xx.
func CheckResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	slurp, err := ioutil.ReadAll(res.Body)
	if err == nil {
		jerr := new(errorReply)
		err = json.Unmarshal(slurp, jerr)
		if err == nil && jerr.Error != nil {
			if jerr.Error.Code == 0 {
				jerr.Error.Code = res.StatusCode
			}
			jerr.Error.Body = string(slurp)
			return jerr.Error
		}
	}
	return &Error{
		Code:   res.StatusCode,
		Body:   string(slurp),
		Header: res.Header,
	}
}

type ApiCall struct {
	Client_             *http.Client
	Ctx_                context.Context
	HttpMethod_         string
	RootUrl_            string
	ServicePath_        string
	Header_             http.Header
	ETag_               string
	PathParams_         map[string]string
	QueryParams_        url.Values
	BodyParams_         interface{}
	ResponseStatusCode_ int
	ResponseHeader_     http.Header
	ResponseData_       interface{}
}

func NewApiCall(client *http.Client, rootUrl string,servicePath string)*ApiCall {
	c := &ApiCall{}
	c.Client_ = client
	c.RootUrl_ = rootUrl
	c.ServicePath_ = servicePath
	c.Header_ = make(http.Header)
	c.QueryParams_ = make(url.Values)

	return c
}

func (c *ApiCall)SendRequest_()(err error) {
	urls := c.RootUrl_ + "/" + c.ServicePath_
	queryString := c.QueryParams_.Encode()
	if queryString != "" {
		urls += "?" + queryString
	}

	var body *bytes.Buffer = nil
	if c.BodyParams_ != nil {
		body = bytes.NewBuffer(nil)
		err = json.NewEncoder(body).Encode(c.BodyParams_)
		if err != nil {
			return err
		}
	}

	req, _ := http.NewRequest(c.HttpMethod_, urls, body)
	if c.Ctx_ != nil {
		req = req.WithContext(c.Ctx_)
	}

	if c.PathParams_ != nil {
		escaped, unescaped, err := Expand(req.URL.Path, c.PathParams_)
		if err != nil {
			return err
		}
		req.URL.Path = unescaped
		req.URL.RawPath = escaped
	}

	if c.ETag_ != "" {
		req.Header.Set("X-ETAG", c.ETag_)
	}

	for k, v := range c.Header_ {
		req.Header[k] = v
	}

	resp, err := c.Client_.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotModified {
		if resp.Body != nil {
			resp.Body.Close()
		}

		return &Error{
			Code:   resp.StatusCode,
			Header: resp.Header,
		}
	}

	defer CloseBody(resp)

	if err := CheckResponse(resp); err != nil {
		return err
	}

	if c.ResponseData_ != nil {
		if err = json.NewDecoder(resp.Body).Decode(c.ResponseData_); err != nil {
			return err
		}
	}

	c.ResponseStatusCode_ = resp.StatusCode
	c.ResponseHeader_ = resp.Header

	return nil
}