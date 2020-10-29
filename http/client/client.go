package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_uuid "github.com/google/uuid"
)

const (
	defaultMinBackOff = 50 * time.Millisecond
	defaultMaxBackOff = 8 * time.Second
)

var (
	ErrRequestNil    = errors.New("httpclient: request cannot be nil")
	ErrRequestURLNil = errors.New("httpclient: request url cannot be empty")

	random = func(min int64, max int64) int64 {
		return rand.New(rand.NewSource(int64(time.Now().Nanosecond()))).Int63n(max-min) + min
	}

	// Decorrelated Exponential Backoff
	// source: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
	decorJitterExponentialBackOff = func(attempt int, min int64, max int64) int64 {
		temp := math.Min(float64(defaultMaxBackOff), math.Pow(float64(2), float64(attempt))*float64(defaultMinBackOff))
		tempSleep := (temp / 2) + float64(random(0, int64(temp/2)))
		return int64(math.Min(float64(max), float64(random(min, int64(tempSleep*3)))))
	}

	once   sync.Once
	client *Client
)

type Client struct {
	*http.Client
}

type retry struct {
	rt         http.RoundTripper
	nums       int
	retry      func(*http.Response, error) bool
	backOff    func(attempt int, minWait time.Duration, maxWait time.Duration) time.Duration
	minBackOff time.Duration
	maxBackOff time.Duration
}

type logRT struct {
	logger *log.Logger
	rt     http.RoundTripper
}

type Opts struct {
	MaxIdleConns    int
	IdleConnTimeout time.Duration

	// Logger enable log for successfull or failed request.
	EnableLogger bool

	// MaxRetry maximum retry for transient errors. Ignored for POST or PATCH as it's not safe.
	MaxRetry int

	// RetryPolicy if MaxRetry is zero then this will be ignored.
	// It can use customized retry policy, if MaxRetry is set and RetryPolicy is nil, then default will be used.
	RetryPolicy func(*http.Response, error) bool

	// BackOffPolicy if MaxRetry is zero then this will be ignored. If RetryPolicy is nil then this will be ignored.
	// It can use customized backoff policy, if MaxRetry is set and RetryPolicy is set and BackOffPolicy is nil, then default will be used.
	BackOffPolicy func(attempt int, minWait time.Duration, maxWait time.Duration) time.Duration
	// MinBackOff minimum wait for backoff. If nil then default will be set.
	MinBackOff *time.Duration
	// MaxBackOff maximum wait for backoff. If nil then default will be set
	MaxBackOff *time.Duration

	// Transport Optional. If you want to specify your own RoundTrip logic. Otherwise will be set to http.DefaultTransport.
	// If MaxRetry is more than zero, then this Transport will be executed after retry RoundTripper.
	Transport http.RoundTripper
}

func newClient(opts *Opts) *Client {
	logger := log.New(os.Stderr, "", 0)
	c := &Client{&http.Client{}}
	if opts.Transport != nil {
		c.Client.Transport = opts.Transport
		if opts.EnableLogger {
			l := &logRT{
				logger: logger,
				rt:     opts.Transport,
			}
			c.Client.Transport = l
		}
		if opts.MaxRetry > 0 {
			transport := opts.Transport
			if opts.EnableLogger {
				l := &logRT{
					logger: logger,
					rt:     opts.Transport,
				}
				transport = l
			}
			re := &retry{
				nums:       opts.MaxRetry,
				rt:         transport,
				retry:      defaultRetry,
				backOff:    defaultBackOff,
				minBackOff: defaultMinBackOff,
				maxBackOff: defaultMaxBackOff,
			}
			if opts.RetryPolicy != nil {
				re.retry = opts.RetryPolicy
			}
			if opts.BackOffPolicy != nil {
				re.backOff = opts.BackOffPolicy
			}
			if opts.MinBackOff != nil {
				re.minBackOff = *opts.MinBackOff
			}
			if opts.MaxBackOff != nil {
				re.maxBackOff = *opts.MaxBackOff
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
		var transport http.RoundTripper = tr
		if opts.MaxRetry > 0 {
			if opts.EnableLogger {
				l := &logRT{
					logger: logger,
					rt:     tr,
				}
				transport = l
			}
			re := &retry{
				nums:       opts.MaxRetry,
				rt:         transport,
				retry:      defaultRetry,
				backOff:    defaultBackOff,
				minBackOff: defaultMinBackOff,
				maxBackOff: defaultMaxBackOff,
			}
			if opts.RetryPolicy != nil {
				re.retry = opts.RetryPolicy
			}
			if opts.BackOffPolicy != nil {
				re.backOff = opts.BackOffPolicy
			}
			if opts.MinBackOff != nil {
				re.minBackOff = *opts.MinBackOff
			}
			if opts.MaxBackOff != nil {
				re.maxBackOff = *opts.MaxBackOff
			}
			c.Client.Transport = re
		} else {
			if opts.EnableLogger {
				l := &logRT{
					logger: logger,
					rt:     tr,
				}
				transport = l
			}
			c.Client.Transport = transport
		}
	}
	return c
}

func New(opts *Opts) *Client {
	c := newClient(opts)
	once.Do(func() {
		client = c
	})
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
	// no retry for non-idempotent methods
	if req.Method == http.MethodPost || req.Method == http.MethodPatch {
		resp, err = r.rt.RoundTrip(req)
	} else {
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
				if cancel != nil {
					cancel()
				}
				return
			}
			if cancel != nil {
				cancel()
			}
			time.Sleep(r.backOff(i, r.minBackOff, r.maxBackOff))
		}
	}
	return
}

func (l *logRT) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = l.rt.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		l.logger.Printf("%s | httpclient | %s | ERR | %s | %v | %v | %s\n", time.Now().Format(time.RFC3339), req.Method, fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path), err, elapsed, req.Header.Get("Request-Id"))
		return
	}
	if resp != nil && resp.StatusCode >= 400 {
		l.logger.Printf("%s | httpclient | %s | %s | %s | %v | %s\n", time.Now().Format(time.RFC3339), req.Method, strconv.Itoa(resp.StatusCode), fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path), elapsed, req.Header.Get("Request-Id"))
		return
	}
	fmt.Printf("%s | httpclient | %s | %s | %s | %v | %s\n", time.Now().Format(time.RFC3339), req.Method, strconv.Itoa(resp.StatusCode), fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path), elapsed, req.Header.Get("Request-Id"))
	return
}

type Request struct {
	BaseURL   string
	Header    http.Header
	URLValues url.Values        // for get method
	Body      map[string]string // for x-www-form-urlencoded, json payload, and multipart/form non-binary data
	Files     []File            // for multipart/form binary data

	RequestID string // unique identifier for each request. E.g. uuid v4. If empty, then will be set automatically using uuid v4.
}

type File struct {
	FieldName string
	FileName  string // along with file extension
	Data      []byte
}

type Response struct {
	StatusCode int
	Status     string
	Header     http.Header
	Body       []byte
	Error      error
}

func (r *Request) init() error {
	if r.RequestID == "" {
		r.RequestID = _uuid.New().String()
	}
	if r.BaseURL == "" {
		return ErrRequestURLNil
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	return nil
}

// URLQuery create new url along with query string from URL values.
func (r *Request) URLQuery() (string, error) {
	if err := r.init(); err != nil {
		return "", err
	}
	urlWithValues, err := url.Parse(r.BaseURL)
	if err != nil {
		return "", err
	}
	uv := urlWithValues.Query()
	for k, v := range r.URLValues {
		for _, e := range v {
			uv.Add(k, e)
		}
	}
	urlWithValues.RawQuery = uv.Encode()
	return urlWithValues.String(), nil
}

// FormURLEncoded create url encoded from Body.
func (r *Request) FormURLEncoded() (io.Reader, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	body := url.Values{}
	for key, val := range r.Body {
		body.Add(key, val)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return strings.NewReader(body.Encode()), nil
}

// JSON create application/json payload from body
func (r *Request) JSON() (io.Reader, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	buf, err := json.Marshal(r.Body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	return bytes.NewBuffer(buf), nil
}

// MultipartForm create multipart/form non-binary and binary data
func (r *Request) MultipartForm() (io.Reader, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
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
		part, err := writer.CreateFormFile(v.FieldName, v.FileName)
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

// Scan fetch r.Body into destination in form of structure or type.
func (r *Response) Scan(destination interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	return json.Unmarshal(r.Body, destination)
}

// String convert response body to string.
func (r *Response) String() (string, error) {
	if r.Error != nil {
		return "", r.Error
	}
	var b strings.Builder
	if _, err := b.Write(r.Body); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Err return response error.
func (r *Response) Err() error {
	return r.Error
}

func (c *Client) do(ctx context.Context, method string, urlQuery string, header http.Header, body io.Reader) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, urlQuery, body)
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Date", time.Now().Format(time.RFC1123))
	return c.Do(req)
}

func (c *Client) call(ctx context.Context, method string, req *Request, body io.Reader) *Response {
	if req == nil {
		return &Response{Error: ErrRequestNil}
	}
	urlQuery, err := req.URLQuery()
	if err != nil {
		return &Response{Error: err}
	}
	req.Header.Set("Request-Id", req.RequestID)
	resp, err := c.do(ctx, method, urlQuery, req.Header, body)
	if err != nil {
		return &Response{Error: err}
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Response{Error: err}
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       respBody,
	}
}

func (c *Client) Get(ctx context.Context, req *Request) *Response {
	return c.call(ctx, http.MethodGet, req, nil)
}

func (c *Client) Head(ctx context.Context, req *Request) *Response {
	return c.call(ctx, http.MethodHead, req, nil)
}

func (c *Client) Options(ctx context.Context, req *Request) *Response {
	return c.call(ctx, http.MethodOptions, req, nil)
}

func (c *Client) PostJSON(ctx context.Context, req *Request) *Response {
	body, err := req.JSON()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPost, req, body)
}

func (c *Client) PostForm(ctx context.Context, req *Request) *Response {
	body, err := req.FormURLEncoded()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPost, req, body)
}

func (c *Client) PostMultipart(ctx context.Context, req *Request) *Response {
	body, err := req.MultipartForm()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPost, req, body)
}

func (c *Client) PutJSON(ctx context.Context, req *Request) *Response {
	body, err := req.JSON()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPut, req, body)
}

func (c *Client) PutForm(ctx context.Context, req *Request) *Response {
	body, err := req.FormURLEncoded()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPut, req, body)
}

func (c *Client) PutMultipart(ctx context.Context, req *Request) *Response {
	body, err := req.MultipartForm()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPut, req, body)
}

func (c *Client) PatchJSON(ctx context.Context, req *Request) *Response {
	body, err := req.JSON()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPatch, req, body)
}

func (c *Client) PatchForm(ctx context.Context, req *Request) *Response {
	body, err := req.FormURLEncoded()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPatch, req, body)
}

func (c *Client) PatchMultipart(ctx context.Context, req *Request) *Response {
	body, err := req.MultipartForm()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodPatch, req, body)
}

func (c *Client) Delete(ctx context.Context, req *Request) *Response {
	body, err := req.FormURLEncoded()
	if err != nil {
		return &Response{Error: err}
	}
	return c.call(ctx, http.MethodDelete, req, body)
}

// ---------------------------------------------------
// Instant functions without client object.
// If called without/before initiating client object, then these functions have no logger and retry option.
// ---------------------------------------------------

func Get(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.Get(ctx, req)
}

func Head(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.Head(ctx, req)
}

func Options(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.Options(ctx, req)
}

func PostJSON(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PostJSON(ctx, req)
}

func PostForm(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PostForm(ctx, req)
}

func PostMultipart(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PostMultipart(ctx, req)
}

func PutJSON(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PutJSON(ctx, req)
}

func PutForm(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PutForm(ctx, req)
}

func PutMultipart(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PutMultipart(ctx, req)
}

func PatchJSON(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PatchJSON(ctx, req)
}

func PatchForm(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PatchForm(ctx, req)
}

func PatchMultipart(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.PatchMultipart(ctx, req)
}

func Delete(ctx context.Context, req *Request) *Response {
	once.Do(func() {
		client = newClient(&Opts{
			MaxIdleConns:    100,
			IdleConnTimeout: 30,
		})
	})
	return client.Delete(ctx, req)
}
