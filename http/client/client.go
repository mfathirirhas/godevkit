package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrRequestNil = errors.New("http client: request cannot be nil")
)

type Client struct {
	*http.Client
}

type retry struct {
	rt   http.RoundTripper
	nums int
}

type Options struct {
	MaxIdleConns    int
	IdleConnTimeout time.Duration

	// MaxRetry maximum retries for timeout. If zero then no retry.
	MaxRetry int

	// Transport if you want to specify your own RoundTrip logic. Otherwise will be set to http.DefaultTransport.
	Transport http.RoundTripper
}

func (r *retry) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	for i := 0; i < r.nums; i++ {
		resp, err = r.rt.RoundTrip(req)
		if req.Context().Err() == context.DeadlineExceeded {
			// if timeout then retry
			continue
		}
		// return if no error and resp is not nil
		if resp != nil && err == nil {
			return
		}
	}
	return
}

func New(opts *Options) *Client {
	c := &Client{&http.Client{}}
	if opts.Transport != nil {
		c.Client.Transport = opts.Transport
		if opts.MaxRetry > 0 {
			re := &retry{
				nums: opts.MaxRetry,
				rt:   opts.Transport,
			}
			c.Client.Transport = re
		}
	} else {
		tr := http.DefaultTransport.(*http.Transport)
		if opts.MaxIdleConns > 0 {
			tr.MaxIdleConns = opts.MaxIdleConns
		}
		if opts.IdleConnTimeout > 0 {
			tr.IdleConnTimeout = opts.IdleConnTimeout
		}
		if opts.MaxRetry > 0 {
			re := &retry{
				nums: opts.MaxRetry,
				rt:   tr,
			}
			c.Client.Transport = re
		} else {
			c.Client.Transport = tr
		}
	}
	return c
}

type Request struct {
	BaseURL   string
	Header    http.Header
	URLValues url.Values        // for get method
	Body      map[string]string // for x-www-form-urlencoded and json payload
	Files     []File            // for multipart/form
}

type File struct {
	FieldName string
	Name      string // along with file extension
	Data      []byte
}

type Response struct {
	StatusCode int
	Status     string
	Header     http.Header
	Body       io.ReadCloser
	Error      error
}

// URLQuery create new url along with query string from URL values.
func (r *Request) URLQuery() (string, error) {
	urlWithValues, err := url.Parse(r.BaseURL)
	if err != nil {
		return "", err
	}
	urlWithValues.RawQuery = r.URLValues.Encode()
	return urlWithValues.String(), nil
}

// URLEncoded create url encoded from Body.
func (r *Request) URLEncoded() (io.Reader, error) {
	body := url.Values{}
	for key, val := range r.Body {
		body.Add(key, val)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return strings.NewReader(body.Encode()), nil
}

// JSON create application/json payload from body
func (r *Request) JSON() (io.Reader, error) {
	buf, err := json.Marshal(r.Body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	return bytes.NewBuffer(buf), nil
}

func (r *Request) MultipartForm() (io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// non-file body
	for key, val := range r.Body {
		err := writer.WriteField(key, fmt.Sprintf("%v", val))
		if err != nil {
			return nil, err
		}
	}
	// file body
	for _, v := range r.Files {
		part, err := writer.CreateFormFile(v.FieldName, v.Name)
		if err != nil {
			return nil, err
		}
		_, err = part.Write(v.Data)
		if err != nil {
			return nil, err
		}
	}

	r.Header.Set("Content-Type", writer.FormDataContentType())
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Struct convert response body to struct. Please input reference to the struct.
func (r *Response) Struct(destination interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	return json.NewDecoder(r.Body).Decode(destination)
}

// String convert response body to string.
func (r *Response) String() (string, error) {
	if r.Error != nil {
		return "", r.Error
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Err return response error.
func (r *Response) Err() error {
	return r.Error
}

func (c *Client) do(ctx context.Context, method string, urlQuery string, header http.Header, body io.Reader) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var (
		req *http.Request
		err error
	)
	switch method {
	case http.MethodPost:
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, urlQuery, body)
	default: // default get
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, urlQuery, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Set("Connection", "keep-alive")
	return c.Do(req)
}

func (c *Client) Get(ctx context.Context, req *Request) *Response {
	if req == nil {
		return &Response{Error: ErrRequestNil}
	}
	urlQuery, err := req.URLQuery()
	if err != nil {
		return &Response{Error: err}
	}
	resp, err := c.do(ctx, http.MethodGet, urlQuery, req.Header, nil)
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       resp.Body,
		Error:      nil,
	}
}

func (c *Client) PostJSON(ctx context.Context, req *Request) *Response {
	if req == nil {
		return &Response{Error: ErrRequestNil}
	}
	urlQuery, err := req.URLQuery()
	if err != nil {
		return &Response{Error: err}
	}
	body, err := req.JSON()
	if err != nil {
		return &Response{Error: err}
	}
	resp, err := c.do(ctx, http.MethodPost, urlQuery, req.Header, body)
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       resp.Body,
		Error:      nil,
	}
}

func (c *Client) PostURLEncoded(ctx context.Context, req *Request) *Response {
	if req == nil {
		return &Response{Error: ErrRequestNil}
	}
	urlQuery, err := req.URLQuery()
	if err != nil {
		return &Response{Error: err}
	}
	urlEncoded, err := req.URLEncoded()
	if err != nil {
		return &Response{Error: err}
	}

	resp, err := c.do(ctx, http.MethodPost, urlQuery, req.Header, urlEncoded)
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       resp.Body,
		Error:      nil,
	}
}

func (c *Client) PostForm(ctx context.Context, req *Request) *Response {
	if req == nil {
		return &Response{Error: ErrRequestNil}
	}
	urlQuery, err := req.URLQuery()
	if err != nil {
		return &Response{Error: err}
	}
	multipartFormData, err := req.MultipartForm()
	if err != nil {
		return &Response{Error: err}
	}

	resp, err := c.do(ctx, http.MethodPost, urlQuery, req.Header, multipartFormData)
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       resp.Body,
		Error:      nil,
	}
}
