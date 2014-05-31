package gohttprequest

import (
	_"github.com/mgutz/ansi"
	"bytes"
	"errors"
	_"bufio"
	_"fmt"
	"net/url"
	"net/http"
	"os"
	"strings"
	"io"
	"io/ioutil"
	"strconv"
	"github.com/woanware/goutil"
)

type HttpRequest struct {
	//headers      []Tuple
	headers      	http.Header
	method       	string
	url          	url.URL
	body 			io.ReadCloser
	//QueryString  	interface{}
	//Timeout      	time.Duration
	//ContentType  	string
	//Accept       	string
	host         	string
	//UserAgent    	string
	MaxRedirects 	int
	ContentLength 	int64
}

type Tuple struct {
	name  string
	value string
}

func New() *HttpRequest {
	m := &HttpRequest{}
	m.SetMethod("GET")
	m.AddHeader("Accept", "*.*")
	m.SetPort(80)
	return m
}

func (r *HttpRequest) Host() (string) {
	return r.host
}

func (r *HttpRequest) Path() (string) {
	return r.url.Path
}

func (r *HttpRequest) Port() (int) {
	index := strings.Index(r.url.Host, ":")
	if index > -1 {
		temp := r.url.Host[index + 1:]
		port, err := strconv.Atoi(temp)
		if err == nil {
			return port
		}
		return 80
	} else {
		if r.url.Scheme == "https" {
			return 443
		} else {
			return 80
		}
	}
}

func (r *HttpRequest) Scheme() (string) {
	return r.url.Scheme
}

func (r *HttpRequest) Fragment() (string) {
	return r.url.Fragment
}

func (r *HttpRequest) Headers() (http.Header) {
	return r.headers
}

func (r *HttpRequest) Query() (string) {
	return r.url.Query().Encode()
}

func (r *HttpRequest) Cookies() ([]http.Cookie) {
	cookies := make([]http.Cookie, 0)
	for _, v := range r.Cookies() {
		cookie := http.Cookie{Name: v.Name,Value: v.Value };
		cookies = append(cookies, cookie)
	}

	return cookies
}

func (r *HttpRequest) SetAddress(uri string) (err error) {
	temp, err := url.Parse(uri)
	if err != nil {
		return errors.New("Cannot parse URI")
	}

	r.url = *temp
	r.host = r.url.Host
	return nil
}

func (r *HttpRequest) SetHost(host string) (err error) {
	temp, err := url.Parse(r.url.Scheme + "://" + host + r.url.Path)
	if err != nil {
		return errors.New("Cannot parse URL")
	}

	r.url = *temp
	r.host = r.url.Host
	return nil
}

func (r *HttpRequest) SetPath(path string) (err error) {
	if path[0:1] != "/" { // /
		return errors.New("Invalid URL")
	}

	tempUrl, err := url.Parse(r.url.Scheme + "://" + r.url.Host + path)
	if err != nil {
		return errors.New("Cannot parse URL: " + err.Error())
	}

	r.url = *tempUrl
	return nil
}

func (r *HttpRequest) SetPort(port int) (err error) {
	// Get the current port if defined
	host := ""
	index := strings.Index(r.url.Host, ":")
	if index > -1 {
		host = r.url.Host[0:index]
	} else {
		host = r.url.Host
	}

	temp, err := url.Parse(r.url.Scheme + "://" + host + ":" + strconv.Itoa(port) + r.url.Path)
	if err != nil {
		return errors.New("Cannot parse URL")
	}

	r.url = *temp
	r.host = r.url.Host
	return nil
}

func (r *HttpRequest) SetFragment(fragment string){
	r.url.Fragment = fragment
}

func (r *HttpRequest) SetScheme(scheme string) (err error){
	if scheme != "http" || scheme != "https" {
		return errors.New("Invalid scheme")
	}

	r.url.Scheme = scheme
	return nil
}

func (r *HttpRequest) SetMethod(method string) (err error){
	httpMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
	index := goutil.GetStringSlicePosition(httpMethods, method)
	if index == -1 {
		return errors.New("Unsupported HTTP method: " + method)
	}

	r.method = method
	return nil
}

func (r *HttpRequest) SetBodyAsString(body string) (){
	r.body = goutil.NopCloser{strings.NewReader(body)}
	var i64 int64
	i64 = int64(len(body))
	r.ContentLength = i64
}

func (r *HttpRequest) SetBodyAsBytes(body []byte) (){
	r.body = goutil.NopCloser{bytes.NewBuffer(body)}
	var i64 int64
	i64 = int64(len(body))
	r.ContentLength = i64
}

func (r *HttpRequest) Body() (string) {
	if r.body == nil {
		return string("")
	}

	body, _ := ioutil.ReadAll(r.body)
	if len(body) > 0 {
		r.body = goutil.NopCloser{bytes.NewBuffer(body)}
		return string(body)
	} else {
		return string("")
	}
}

func (r *HttpRequest) AddHeader(name string, value string) {
	if r.headers == nil {
		r.headers = make(http.Header)
	}

	ret := r.headers.Get(name)
	if len(ret) == 0 {
		r.headers.Add(name, value)
	} else {
		r.headers.Set(name, value)
	}
}

func (r *HttpRequest) AddQuery(name string, value string) (err error) {
	values := r.url.Query()
	values.Add(name, value)
	r.url.RawQuery = values.Encode()
	return nil
}

func (r *HttpRequest) AddCookie(name string, value string) (err error) {
	// Get a list of the existing cookies since we cannot directly modify the ones already assigned to the http
	// struct. If we identify the cookie that we are trying to add/modify then we don't include it, since we will
	// need to recreate it with the new values
	cookies := []http.Cookie{}
	for _, v := range r.Cookies() {
		if v.Name != name {
			cookie := http.Cookie{Name: v.Name,Value: v.Value };
			cookies = append(cookies, cookie)
		}
	}

	// Delete the cookie header
	r.headers.Del("Cookie")

	// Now add the new/modified cookie to the http struct
	cookie := http.Cookie{Name: name,Value: value };
	cookies = append(cookies, cookie)

	// Add the cookies to the http struct
	for _, v := range cookies {
		r.AddCookie(v.Name, v.Value)
	}
	return nil
}

func (r *HttpRequest) RemoveHeader(header string) {
	ret := r.Headers().Get(header)
	if len(ret) > 0 {
		r.Headers().Del(header)
	}
}

func (r *HttpRequest) RemoveCookie(cookie string) {
	cookies := []http.Cookie{}
	for _, v := range r.Cookies() {
		if v.Name != cookie {
			cookie := http.Cookie{Name: v.Name,Value: v.Value };
			cookies = append(cookies, cookie)
		}
	}

	// Delete the cookie header
	r.headers.Del("Cookie")

	// Add the cookies to the http struct
	for _, v := range cookies {
		r.AddCookie(v.Name, v.Value)
	}
}

func (r *HttpRequest) RemoveQuery(query string) {
	ret := r.url.Query().Get(query)
	if len(ret) > 0 {
		r.url.Query().Del(query)
	}
}

func (r *HttpRequest) ClearHeaders() () {
	for k := range r.Headers() {
		delete(r.Headers(), k)
	}
}

func (r *HttpRequest) ClearCookies() () {
	header := r.headers.Get("Cookie")
	if len(header) > 0 {
		r.headers.Del("Cookie")
	}
}

func (r *HttpRequest) ClearFragment() () {
	r.url.Fragment = ""
}

func (r *HttpRequest) ClearQuery() () {
	for k := range r.url.Query() {
		delete(r.url.Query(), k)
	}
}

func (r *HttpRequest) ClearBody() () {
	r.body = goutil.NopCloser{strings.NewReader("")}
}

// ##### Helper Methods ################################################################################################

// Retrieve the HTTP/HTTPS proxy from the environment
func GetProxy(scheme string) *url.URL {
	var proxy_url *url.URL
	if scheme == "https" {
		temp := os.Getenv("https_proxy")
		if len(temp) > 0 {
			proxy_url, _ = url.Parse(temp)
		} else {
			temp := os.Getenv("HTTPS_PROXY")
			if len(temp) > 0 {
				proxy_url, _ = url.Parse(temp)
			}
		}
	} else {
		temp := os.Getenv("http_proxy")
		if len(temp) > 0 {
			proxy_url, _ = url.Parse(temp)
		} else {
			temp := os.Getenv("HTTP_PROXY")
			if len(temp) > 0 {
				proxy_url, _ = url.Parse(temp)
			}
		}
	}

	return proxy_url
}

// Custom function to ensure that no redirects occur
func NoRedirects(req *http.Request, via []*http.Request) error {
	return errors.New("No redirects allowed")
}

// Custom function to validate the HTTP redirects
func AllowRedirects(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

// ##### HTTP Methods ##################################################################################################

func (r *HttpRequest) Send() (*http.Response, error) {
	if r.method == "" {
		r.method = "GET"
	}

	// Retrieve the proxy details (if required)
	proxy_url := GetProxy(r.url.Scheme)

	// Define the http client, with the custom method for redirect validation
	var httpClient *http.Client
	if r.MaxRedirects == 0 {
		httpClient = &http.Client{Transport: &http.Transport {Proxy: http.ProxyURL(proxy_url)}, CheckRedirect: NoRedirects}
	} else {
		httpClient = &http.Client{Transport: &http.Transport {Proxy: http.ProxyURL(proxy_url)}, CheckRedirect: AllowRedirects}
	}

	var response *http.Response
	var err error
	switch r.method {
	case "GET", "HEAD", "DELETE", "OPTIONS":
		response, err = r.doHttpRequest(httpClient)
	case "POST", "PUT":
		response, err = r.doHttpRequestWithBody(httpClient)
	}

	return response, err
}

func (r *HttpRequest) doHttpRequest(httpClient *http.Client) (*http.Response, error){
	request, err := http.NewRequest(r.method, r.url.String(), nil)
	request.Header = r.headers

	//	if len(user) > 0 {
	//		request.SetBasicAuth(user, password)
	//	}

	response, err := httpClient.Do(request)
	return response, err
}

func  (r *HttpRequest) doHttpRequestWithBody(httpClient *http.Client) (*http.Response, error) {

	request, err := http.NewRequest(r.method, r.url.String(), strings.NewReader(r.Body()))
	request.Header = r.headers

	//	if len(user) > 0 {
	//		request.SetBasicAuth(user, password)
	//	}
	//

	response, err := httpClient.Do(request)
	return response, err
}
