package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultMinBackOff = 50 * time.Millisecond
	defaultMaxBackOff = 8 * time.Second
)

var (
	ErrRequestNil = errors.New("http client: request cannot be nil")

	random = func(min int64, max int64) int64 {
		return rand.New(rand.NewSource(int64(time.Now().Nanosecond()))).Int63n(max-min) + min
	}

	// Decorrelated Exponential Backoff
	// source: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/ &&
	decorJitterExponentialBackOff = func(attempt int, min int64, max int64) int64 {
		temp := math.Min(float64(defaultMaxBackOff), math.Pow(float64(2), float64(attempt))*float64(defaultMinBackOff))
		tempSleep := (temp / 2) + float64(random(0, int64(temp/2)))
		return int64(math.Min(float64(max), float64(random(min, int64(tempSleep*3)))))
	}
)

type Client struct {
	*http.Client
}

type retry struct {
	rt      http.RoundTripper
	nums    int
	retry   func(*http.Response, error) bool
	backOff func(attempt int, minWait time.Duration, maxWait time.Duration) time.Duration
}

type Options struct {
	MaxIdleConns    int
	IdleConnTimeout time.Duration

	// MaxRetry maximum retries for timeout. If zero then no retry.
	MaxRetry int

	// RetryPolicy if MaxRetry is zero then this will be ignored.
	// It can use customized retry policy, if MaxRetry is set and RetryPolicy is nil, then default will be used.
	RetryPolicy func(*http.Response, error) bool

	// BackOffPolicy if MaxRetry is zero then this will be ignored. If RetryPolicy is nil then this will be ignored.
	// It can use customized backoff policy, if MaxRetry is set and RetryPolicy is set and BackOffPolicy is nil, then default will be used.
	BackOffPolicy func(attempt int, minWait time.Duration, maxWait time.Duration) time.Duration

	// Transport Optional. If you want to specify your own RoundTrip logic. Otherwise will be set to http.DefaultTransport.
	Transport http.RoundTripper
}

func New(opts *Options) *Client {
	c := &Client{&http.Client{}}
	if opts.Transport != nil {
		c.Client.Transport = opts.Transport
		if opts.MaxRetry > 0 {
			re := &retry{
				nums:    opts.MaxRetry,
				rt:      opts.Transport,
				retry:   defaultRetry,
				backOff: defaultBackOff,
			}
			if opts.RetryPolicy != nil {
				re.retry = opts.RetryPolicy
			}
			if opts.BackOffPolicy != nil {
				re.backOff = opts.BackOffPolicy
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
				nums:    opts.MaxRetry,
				rt:      tr,
				retry:   defaultRetry,
				backOff: defaultBackOff,
			}
			if opts.RetryPolicy != nil {
				re.retry = opts.RetryPolicy
			}
			if opts.BackOffPolicy != nil {
				re.backOff = opts.BackOffPolicy
			}
			c.Client.Transport = re
		} else {
			c.Client.Transport = tr
		}
	}
	return c
}

// defaultRetry retry policy
func defaultRetry(resp *http.Response, err error) bool {
	if resp != nil {
		// transient http status codes
		if resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			return true
		}
	}

	if err != nil {
		if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
			return true
		}
	}

	return false
}

// defaultBackOff default back off
func defaultBackOff(attempt int, min time.Duration, max time.Duration) time.Duration {
	return time.Duration(decorJitterExponentialBackOff(attempt, int64(min), int64(max)))
}

func (r *retry) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var (
		duration time.Duration
		ctx      context.Context
		cancel   func()
	)
	if deadline, ok := req.Context().Deadline(); ok {
		duration = time.Until(deadline)
	}
	for i := 0; i < r.nums; i++ {
		if duration > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), duration)
			req = req.WithContext(ctx)
		}
		resp, err = r.rt.RoundTrip(req)
		if !r.retry(resp, err) {
			cancel()
			return
		}
		cancel()
		time.Sleep(r.backOff(i, defaultMinBackOff, defaultMaxBackOff))
	}
	return
}

type Request struct {
	URL       string
	Header    http.Header
	URLValues url.Values        // for get method
	Body      map[string]string // for x-www-form-urlencoded, json payload, and multipart/form non-binary data
	Files     []File            // for multipart/form binary data
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
	urlWithValues, err := url.Parse(r.URL)
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

// MultipartForm create multipart/form non-binary and binary data
func (r *Request) MultipartForm() (io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// non-binary body
	for key, val := range r.Body {
		err := writer.WriteField(key, fmt.Sprintf("%v", val))
		if err != nil {
			return nil, err
		}
	}
	// binary body
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
	if r.Body != nil {
		defer r.Body.Close()
	}
	if r.Error != nil {
		return r.Error
	}
	return json.NewDecoder(r.Body).Decode(destination)
}

// String convert response body to string.
func (r *Response) String() (string, error) {
	if r.Body != nil {
		defer r.Body.Close()
	}
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
	if r.Body != nil {
		defer r.Body.Close()
	}
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
