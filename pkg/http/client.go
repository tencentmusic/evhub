package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// Method is name of method
type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

var (
	jsonCheck = regexp.MustCompile("(?i:(?:application|text)/json)")
)

// ClientConfig is configuration http server
type ClientConfig struct {
}

// Client is http client
type Client struct {
	Client *http.Client
}

// NewClient creates a new http client
func NewClient(c *ClientConfig) *Client {
	return &Client{Client: &http.Client{}}
}

// PrepareRequest build the request
func (c *Client) PrepareRequest(
	ctx context.Context,
	path string, method Method,
	postBody interface{},
	headerParams map[string]string,
	queryParams url.Values) (request *http.Request, err error) {
	var body *bytes.Buffer

	if headerParams == nil {
		headerParams = make(map[string]string)
	}
	// Detect postBody type and post.
	if postBody != nil {
		var contentType = headerParams["Content-Type"]
		if contentType == "" {
			contentType = detectContentType(postBody)
			headerParams["Content-Type"] = contentType
		}
		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}

	// Setup path and query parameters
	urlPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Adding Query Param
	query := urlPath.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	urlPath.RawQuery = query.Encode()

	// Generate a new request
	if body != nil {
		request, err = http.NewRequest(string(method), urlPath.String(), body)
	} else {
		request, err = http.NewRequest(string(method), urlPath.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers.Set(h, v)
		}
		request.Header = headers
	}

	// add context to the request, if not nil
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	return request, nil
}

// DoProcess
func (c *Client) DoProcess(request *http.Request, response interface{}) error {
	if response == nil {
		return errors.New("response must be non-nil")
	}

	// do request
	resp, err := c.Client.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http res:code = %d, status:%s", resp.StatusCode, resp.Status))
	}

	localVarBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return decode(response, localVarBody, resp.Header.Get("Content-Type"))
}

// detectContentType method is used to figure out `Request.Body` content type for request header
func detectContentType(body interface{}) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Ptr:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

func setBody(body interface{}, contentType string) (bodyBuf *bytes.Buffer, err error) {
	bodyBuf = &bytes.Buffer{}
	if reader, ok := body.(io.Reader); ok {
		_, err = bodyBuf.ReadFrom(reader)
	} else if b, ok := body.([]byte); ok {
		_, err = bodyBuf.Write(b)
	} else if s, ok := body.(string); ok {
		_, err = bodyBuf.WriteString(s)
	} else if s, ok := body.(*string); ok {
		_, err = bodyBuf.WriteString(*s)
	} else if jsonCheck.MatchString(contentType) {
		err = json.NewEncoder(bodyBuf).Encode(body)
	}

	if err != nil {
		return nil, err
	}

	if bodyBuf.Len() == 0 {
		err = errors.New(fmt.Sprintf("Invalid body type %s\n", contentType))
		return nil, err
	}
	return bodyBuf, nil
}

func decode(v interface{}, b []byte, contentType string) error {
	if strings.Contains(contentType, "application/xml") {
		dec := json.NewDecoder(bytes.NewReader(b))
		if err := dec.Decode(v); err == nil {
			return nil
		} else {
			return errors.New("json decode error")
		}
	}
	return errors.New("undefined response type")
}
