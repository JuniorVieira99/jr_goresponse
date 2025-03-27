// Package response provides functionality for handling HTTP responses, including creating,
// parsing, and managing collections of responses.
//
// The package includes:
// - Response: A struct representing a single HTTP response with methods for manipulation.
// - ResponsePack: A thread-safe collection of Response objects with statistics tracking.
// - CompressResponsePack: A memory-efficient collection of compressed Response objects.
//
// These components are designed to work with the jr_httpcodes/codes package for HTTP methods
// and status codes, making it easy to work with HTTP responses in a standardized way.
//
// ResponsePack and CompressResponsePack provide concurrent-safe operations for adding,
// retrieving, and managing sets of Response objects, with CompressResponsePack offering
// lower memory usage through gzip compression.
//
// Basic usage examples:
//
//	// Create a single response
//	resp, err := response.NewResponse(
//	    "https://example.com/api",
//	    "example.com",
//	    codes.GET,
//	    codes.OK,
//	    map[string]string{"Content-Type": "application/json"},
//	    []byte(`{"status":"success"}`),
//	    0,
//	    nil,
//	)
//
//	// Create a response pack and add responses
//	pack := response.NewResponsePack()
//	pack.AddResponse(resp)
//
//	// Retrieve and print information
//	retrievedResp := pack.GetResponse("https://example.com/api")
//	retrievedResp.Print()
//
//	// Parse raw HTTP response
//	rawData := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html>...</html>")
//	parsedResp, err := response.ParseRawHTTPResponse(&rawData, "https://example.com")
package response

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlPack "net/url"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/JuniorVieira99/jr_httpcodes/codes"
)

// Response struct
// ----------------------------------------------------------------------

// Response struct
type Response struct {
	Method      codes.Method      `json:"method"`
	StatusCode  codes.StatusCode  `json:"statusCode"`
	Url         string            `json:"url"`
	Host        string            `json:"host"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	BodyLength  uint64            `json:"bodyLength"`
	RawResponse []byte            `json:"rawResponse"`
}

type ConfigResponse struct {
	Method      codes.Method
	StatusCode  codes.StatusCode
	Url         string
	Host        string
	Headers     map[string]string
	Body        []byte
	BodyLength  uint64
	RawResponse []byte
}

// ToString returns a string representation of the Response object, including
// URL, host, method, status code, headers, body, and body length.
func (r *Response) ToString() string {
	var sb strings.Builder

	sb.Grow(256) // Pre-allocate memory for better performance

	sb.WriteString("\nUrl: ")
	sb.WriteString(r.Url)
	sb.WriteString("\nHost: ")
	sb.WriteString(r.Host)
	sb.WriteString("\nMethod: ")
	sb.WriteString(r.Method.String())
	sb.WriteString("\nStatusCode: ")
	sb.WriteString(fmt.Sprintf("%d", r.StatusCode))

	sb.WriteString("\nHeaders:")
	for key, value := range r.Headers {
		sb.WriteString("\n")
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
	}

	// Write the body
	if len(r.Body) > 0 {
		sb.WriteString("\nBody:")
		sb.WriteString(r.ReadBody())
	}

	sb.WriteString(fmt.Sprintf("\nBodyLength: %d", r.BodyLength))
	return sb.String()
}

// ReadBody returns the response body as a string.
func (r *Response) ReadBody() string {
	return string(r.Body)
}

// IsSuccessful checks if the response status code indicates a successful HTTP response.
// It returns true if the status code is within the range of success codes, otherwise false.
func (r *Response) IsSuccessful() bool {
	return codes.IsSuccess(r.StatusCode)
}

// DetailedStatusCode returns the StatusCode as a human-readable string, e.g. "OK" instead of 200.
func (r *Response) DetailedStatusCode() string {
	return r.StatusCode.String()
}

// ReadRawResponse returns the raw response data as a string.
func (r *Response) ReadRawResponse() string {
	return string(r.RawResponse)
}

// Print prints a string representation of the Response struct to the console.
func (r *Response) Print() {
	fmt.Println(r.ToString())
}

// isTextContent checks if the headers indicate text content
func isTextContent(headers map[string]string) bool {
	contentType, exists := headers["Content-Type"]
	if !exists {
		return false
	}

	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-www-form-urlencoded",
	}

	for _, textType := range textTypes {
		if strings.Contains(strings.ToLower(contentType), textType) {
			return true
		}
	}
	return false
}

// ToReadableJSON converts the Response struct to a JSON-encoded byte slice, while
// taking care to properly encode binary data. If the response body or raw response
// contains non-UTF8 data, it will be base64-encoded and the resulting JSON will
// contain an "encoding" section with information about the used encoding.
func (r *Response) ToReadableJSON() ([]byte, error) {

	// Try to convert body to a readable string first
	var bodyContent string
	if isTextContent(r.Headers) && utf8.Valid(r.Body) {
		bodyContent = string(r.Body)
	} else {
		// Fall back to base64
		bodyContent = base64.StdEncoding.EncodeToString(r.Body)
	}

	// Try to convert rawResponse to a readable string first
	var rawResponseContent string
	if utf8.Valid(r.RawResponse) {
		rawResponseContent = string(r.RawResponse)
	} else {
		// Fall back to base64
		rawResponseContent = base64.StdEncoding.EncodeToString(r.RawResponse)
	}

	// Create a temporary struct to handle encoded binary data
	tempData := struct {
		Method      codes.Method      `json:"method"`
		StatusCode  codes.StatusCode  `json:"statusCode"`
		Url         string            `json:"url"`
		Host        string            `json:"host"`
		Headers     map[string]string `json:"headers"`
		Body        string            `json:"body"`
		BodyLength  uint64            `json:"bodyLength"`
		RawResponse string            `json:"rawResponse"`
		Encoding    struct {
			Body        string `json:"body,omitempty"`
			RawResponse string `json:"rawResponse,omitempty"`
		} `json:"encoding,omitempty"`
	}{
		Method:      r.Method,
		StatusCode:  r.StatusCode,
		Url:         r.Url,
		Host:        r.Host,
		Headers:     r.Headers,
		Body:        bodyContent,
		BodyLength:  r.BodyLength,
		RawResponse: rawResponseContent,
	}

	// Add encoding information if we used base64
	if !utf8.Valid(r.Body) {
		tempData.Encoding.Body = "base64"
	}
	if !utf8.Valid(r.RawResponse) {
		tempData.Encoding.RawResponse = "base64"
	}

	jsonData, err := json.Marshal(tempData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// ToJSON converts the Response struct to a JSON-encoded byte slice.
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// Constructors
// ----------------------------------------------------------------------

// NewResponseFromJSON takes a JSON-encoded byte slice and attempts to decode it
// into a Response struct. It returns a pointer to the Response struct and a possible
// error if the decoding fails.
func NewResponseFromJSON(data []byte) (*Response, error) {
	if data == nil {
		return nil, fmt.Errorf("input JSON data is nil")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("input JSON data is empty")
	}

	var outputResponse Response

	err := json.Unmarshal(data, &outputResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON into Response: %w", err)
	}

	return &outputResponse, nil

}

// NewResponse creates a new Response instance with the given parameters and
// returns it with a possible error. It validates the given method and status
// code, and sets default values for the headers and body if they are nil.
func NewResponse(
	url string,
	host string,
	method codes.Method,
	statusCode codes.StatusCode,
	headers map[string]string,
	body []byte,
	bodyLength uint64,
	rawResponse []byte,
) (*Response, error) {
	err := codes.ValidateStatusCode(statusCode)
	if err != nil {
		return nil, err
	}

	err = codes.ValidateMethod(method)
	if err != nil {
		return nil, err
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	if body == nil {
		body = []byte{}
	}

	if bodyLength == 0 {
		bodyLength = uint64(len(body))
	}

	return &Response{
		Method:      method,
		StatusCode:  statusCode,
		Url:         url,
		Host:        host,
		Headers:     headers,
		Body:        body,
		BodyLength:  bodyLength,
		RawResponse: rawResponse,
	}, nil
}

// NewResponseFromConfig creates a new Response instance using the provided ResponseConfig.
// It returns a pointer to the Response struct and a possible error if the Response cannot
// be created. This function leverages the NewResponse function to perform validation and
// initialization of the Response fields.
func NewResponseFromConfig(config ConfigResponse) (*Response, error) {
	return NewResponse(config.Url, config.Host, config.Method, config.StatusCode, config.Headers, config.Body, config.BodyLength, config.RawResponse)
}

// Response Pack
// ----------------------------------------------------------------------

// ResponsePack A struct with many Responses objects
type ResponsePack struct {
	Responses    map[string]map[string]*Response `json:"responses"` // map[URL][round]Response
	Total        uint64                          `json:"total"`
	Success      uint64                          `json:"success"`
	Failure      uint64                          `json:"failure"`
	SuccessRatio float64                         `json:"successRatio"`
	FailureRatio float64                         `json:"failureRatio"`
	Info         map[string]string               `json:"info"`
	mu           sync.RWMutex
}

// GetResponse takes a URL and retrieves a slice of Response objects from the ResponsePack.
// If a response does not exist for a URL, it returns nil.
func (r *ResponsePack) GetResponse(url string) ([]*Response, error) {

	// Output slice
	var resultSlice []*Response

	// Read Map
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, ok := r.Responses[url]

	// If not found
	if !ok {
		return nil, fmt.Errorf("response not found for URL: %s", url)
	}

	// Convert map to slice
	for _, value := range result {
		resultSlice = append(resultSlice, value)
	}

	return resultSlice, nil
}

// BatchGetResponse retrieves the Response objects associated with a given slice of URLs
// from the ResponsePack concurrently. It handles both direct URL lookups and URLs with
// round suffixes, returning a map where the outer key is the URL, and the inner map has
// keys representing the round number (e.g., "round_1") and values as the Response objects.
// The function also returns a slice of errors if any GetResponse operation fails.
func (r *ResponsePack) BatchGetResponse(urls []string) (map[string]map[string]*Response, []error) {
	// Output map
	output := map[string]map[string]*Response{}
	// Output errors
	var errors []error

	// Channels
	errCh := make(chan error, len(urls))
	resultCh := make(chan []*Response, len(urls))

	// WaitGroup
	var wg sync.WaitGroup

	// Process each URL just once
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			// Get response
			mapRes, err := r.GetResponse(url)
			if err != nil {
				errCh <- err
			}
			if mapRes != nil {
				resultCh <- mapRes
			}
		}(url)
	}

	wg.Wait()
	close(errCh)
	close(resultCh)

	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) == 0 {
		errors = nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for result := range resultCh {
		urlKey := result[0].Url
		output[urlKey] = map[string]*Response{}

		for index, response := range result {
			newKey := fmt.Sprintf("round_%d", index)
			output[urlKey][newKey] = response
		}
	}

	return output, errors
}

// GetKeysOfResponses returns a slice containing all the keys present in the Responses map of the ResponsePack.
// It acquires a read lock on the mutex to ensure thread-safe access to the map.
func (p *ResponsePack) GetKeysOfResponses() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.Responses))
	for key := range p.Responses {
		keys = append(keys, key)
	}

	return keys
}

// AddResponse adds a Response object to the ResponsePack struct.
func (p *ResponsePack) AddResponse(response *Response) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	ok := codes.IsSuccess(response.StatusCode)

	if ok {
		p.Success++
	} else {
		p.Failure++
	}
	p.Total++

	// Recalculate ratios directly after updating metrics
	if p.Total > 0 {
		if p.Success != 0 {
			p.SuccessRatio = float64(p.Success) / float64(p.Total)
		}
		if p.Failure != 0 {
			p.FailureRatio = float64(p.Failure) / float64(p.Total)
		}
	}

	var round int = 0

	// Check if response already exists
	_, ok = p.Responses[response.Url]
	// If url does not exists, create inner map
	if !ok {
		p.Responses[response.Url] = make(map[string]*Response)
		p.Responses[response.Url]["round_1"] = response
		return nil
	}

	// Get round
	for range p.Responses[response.Url] {
		round++
	}
	// Make new key
	newKey := fmt.Sprintf("round_%d", round+1)
	// Add response with new key
	p.Responses[response.Url][newKey] = response
	return nil
}

// BatchAddResponse adds a slice of Response objects to the ResponsePack struct,
// handling duplicate URL entries by appending a round suffix.
// It returns a slice of errors if any of the AddResponse operations fail.
func (p *ResponsePack) BatchAddResponse(responses []*Response) []error {
	errCh := make(chan error, len(responses))
	errSlice := make([]error, 0)

	maxWorkers := runtime.NumCPU()
	if len(responses) < maxWorkers {
		maxWorkers = len(responses)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(response *Response) {
			defer wg.Done()
			err := p.AddResponse(response)
			if err != nil {
				errCh <- err
			}
		}(responses[i])
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		errSlice = append(errSlice, err)
	}

	if len(errSlice) > 0 {
		return errSlice
	}

	return nil
}

// Calculate recalculates the success and failure ratios of the ResponsePack.
func (p *ResponsePack) Calculate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Total == 0 {
		return
	}
	if p.Success != 0 {
		p.SuccessRatio = float64(p.Success) / float64(p.Total)
	}
	if p.Failure != 0 {
		p.FailureRatio = float64(p.Failure) / float64(p.Total)
	}
}

// AddInfo adds a key-value pair to the info map of the ResponsePack struct.
func (p *ResponsePack) AddInfo(key string, value string) {
	p.Info[key] = value
}

// ToString returns a string representation of the ResponsePack struct,
// including the total number of responses, number of successful responses,
// number of failed responses, and the success and failure ratios. It also includes the info map as key-value pairs.
func (p *ResponsePack) ToString() string {

	var str strings.Builder
	str.Grow(256)

	p.mu.RLock()
	defer p.mu.RUnlock()

	str.WriteString(fmt.Sprintf("Total: %d", p.Total))
	str.WriteString(fmt.Sprintf("\nSuccess: %d", p.Success))
	str.WriteString(fmt.Sprintf("\nFailure: %d", p.Failure))
	str.WriteString(fmt.Sprintf("\nSuccessRatio: %f", p.SuccessRatio))
	str.WriteString(fmt.Sprintf("\nFailureRatio: %f", p.FailureRatio))
	str.WriteString("\nInfo:")

	for key, value := range p.Info {
		str.WriteString(fmt.Sprintf("\n\t%s: %s", key, value))
	}

	return str.String()
}

// Print prints a string representation of the ResponsePack struct to the console.
func (p *ResponsePack) Print() {
	fmt.Println(p.ToString())
}

// Len returns the number of responses stored in the ResponsePack.
func (p *ResponsePack) Len() int {
	if p == nil {
		return 0
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Responses)
}

// GetErrorReport returns a map of strings to strings, where each key is a URL and
// each value is the status code of the response. The map only contains URLs for
// which the status code was not successful.
//
// The function will return an error if the ResponsePack is nil or if there are no
// responses stored in the pack.
//
// Note: The function acquires a read lock on the ResponsePack's mutex to ensure
// thread-safe access to the Responses map.
func (p *ResponsePack) GetErrorReport() (map[string]map[string]*Response, error) {

	if p == nil {
		return nil, fmt.Errorf("response pack is nil")
	}
	if p.Len() == 0 {
		return nil, fmt.Errorf("no responses found")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	output := map[string]map[string]*Response{}

	for outKey, outValue := range p.Responses {
		for inKey, inValue := range outValue {
			if !codes.IsSuccess(inValue.StatusCode) {
				output[outKey] = make(map[string]*Response)
				output[outKey][inKey] = inValue
			}
		}
	}

	return output, nil
}

// GetErrorReportString returns a string representation of the error report
// for the ResponsePack. It includes URLs and their corresponding status codes
// for responses that were not successful. The function will return an error
// string if the ResponsePack is nil or if there are no failed responses. It
// locks the mutex for reading to ensure thread-safe access to the Responses map.
func (p *ResponsePack) GetErrorReportString() (string, error) {
	var str strings.Builder
	reportMap, err := p.GetErrorReport()

	if err != nil {
		str.WriteString(err.Error())
		return str.String(), err
	}
	str.WriteString("Error Report:\n")
	for key, value := range reportMap {
		str.WriteString("URL: ")
		str.WriteString(key)
		str.WriteString("\n")
		for inKey, inValue := range value {
			str.WriteString(fmt.Sprintf("\t%s: %d\n", inKey, inValue.StatusCode))
		}
	}

	return str.String(), nil
}

// NewResponsePack returns a new ResponsePack instance with zero values for all fields.
func NewResponsePack() *ResponsePack {
	return &ResponsePack{
		Responses:    map[string]map[string]*Response{}, // keys: Urls, values: map[string]Responses
		Info:         map[string]string{},               // keys: string, values: string
		Total:        0,
		Success:      0,
		Failure:      0,
		SuccessRatio: 0,
		FailureRatio: 0,
		mu:           sync.RWMutex{},
	}
}

// responseParser takes a pointer to a byte slice containing HTTP response data and attempts to parse it into a Response struct.
// It returns a pointer to the Response struct and an error if the parsing fails.
// The function returns an error if the response data is empty.
// The function reads the response body into a byte slice and extracts all headers into a map.
// The function determines the host from the headers, and if not available, from the request object.
// The function converts the status code to a codes.StatusCode and the method to a codes.Method.
// The function gets the URL from the request object if available, or an empty string if not available.
// The function creates a new Response object with the extracted data and returns it with any error that may have occurred.
func responseParser(data *[]byte, url string) (*Response, error) {
	if data == nil || len(*data) == 0 {
		return nil, fmt.Errorf("empty response data")
	}

	// Create a buffer reader from the response data
	reader := bufio.NewReader(bytes.NewReader(*data))

	// Parse the HTTP response using the standard library
	httpResponse, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP response: %w", err)
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	defer httpResponse.Body.Close()

	// Extract all headers into a map
	headers := make(map[string]string)
	for name, values := range httpResponse.Header {
		headers[name] = strings.Join(values, ", ")
	}

	var host string

	// Handle URL parsing based on format
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parsedURL, err := urlPack.Parse(url)
		if err != nil {
			return nil, err
		}
		host = parsedURL.Host
	} else {
		// This is likely an IP:port format
		host = url
	}

	// Get host from request if not available
	if host == "" && httpResponse.Request != nil {
		host = httpResponse.Request.Host
	}

	// Convert status code
	statusCode := codes.StatusCode(httpResponse.StatusCode)

	// Convert method (if request is available)
	var method codes.Method
	if httpResponse.Request != nil {
		method = codes.Method(httpResponse.Request.Method)
	} else {
		method = codes.GET // Default to GET if not available
	}

	// Create the response object
	response, err := NewResponse(
		url,
		host,
		method,
		statusCode,
		headers,
		body,
		uint64(len(body)),
		*data,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create response: %w", err)
	}

	// Close response body
	err = httpResponse.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close response body: %w", err)
	}

	// Return the response
	return response, err
}

// ParseRawHTTPResponse takes a pointer to a byte slice containing HTTP response data and attempts to parse it into a Response struct.
// It returns a pointer to the Response struct and an error if the parsing fails.
// The function returns an error if the response data is empty.
func ParseRawHTTPResponse(rawResponse *[]byte, url string) (*Response, error) {
	return responseParser(rawResponse, url)
}

// ParseStringHTTPResponse takes a string containing HTTP response data and attempts to parse it into a Response struct.
// It returns a pointer to the Response struct and an error if the parsing fails.
// The function returns an error if the response data is empty.
func ParseStringHTTPResponse(rawResponse string, url string) (*Response, error) {
	data := []byte(rawResponse)
	return responseParser(&data, url)
}

// Compress and Decompress Response
// ----------------------------------------------------------------------

// Compress takes the Response object and compresses it using Gzip.
// The function will first convert the Response object to a JSON byte slice
// using the ToJSON method, and then use the standard library's gzip package
// to write the JSON data into a bytes.Buffer. The function returns the
// compressed data as a byte slice, and an error if either the ToJSON or
// gzip.Write operation fails.
func (r *Response) Compress() ([]byte, error) {
	jsonData, err := r.ToJSON()
	if err != nil {
		return nil, err
	}
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)
	_, err = gz.Write(jsonData)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return compressedData.Bytes(), nil
}

// NewResponseFromCompressed creates a Response from compressed data
func NewResponseFromCompressed(compressedData []byte) (*Response, error) {
	// Create a reader for the compressed data
	r, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer r.Close()

	// Read the decompressed data
	jsonData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Unmarshal the JSON data
	var response Response
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// CompressResponsePack
// ----------------------------------------------------------------------

type CompressResponsePack struct {
	CompressedResponses map[string]map[string][]byte
	MetaInfo            map[string]string
	mu                  sync.RWMutex
}

// NewCompressResponsePack creates a new CompressResponsePack, initializing the CompressedResponses sync.Map.
func NewCompressResponsePack() *CompressResponsePack {
	return &CompressResponsePack{
		CompressedResponses: make(map[string]map[string][]byte),
		MetaInfo:            make(map[string]string),
		mu:                  sync.RWMutex{},
	}
}

// AddResponse compresses the given Response object and adds it to the CompressedResponses map,
// handling duplicate URL entries by appending a round suffix. It returns an error if compression fails.
func (r *CompressResponsePack) AddResponse(response *Response) error {

	if response == nil {
		return fmt.Errorf("response is nil")
	}

	if r == nil {
		return fmt.Errorf("response pack is nil")
	}

	compressedData, err := response.Compress()
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the URL already exists in the map
	_, ok := r.CompressedResponses[response.Url]

	if !ok {
		// If not, create a new map for the URL
		r.CompressedResponses[response.Url] = map[string][]byte{}
		r.CompressedResponses[response.Url]["round_1"] = compressedData
		return nil
	}

	// If the URL already exists, append a round suffix
	var round int = 0
	for range r.CompressedResponses[response.Url] {
		round++
	}
	r.CompressedResponses[response.Url][fmt.Sprintf("round_%d", round+1)] = compressedData

	return nil
}

// BatchAddResponse adds a slice of Response objects to the CompressResponsePack, compressing each one and handling duplicate URL entries by appending a round suffix.
// It returns a slice of errors if any of the AddResponse operations fail.
func (r *CompressResponsePack) BatchAddResponse(responses []*Response) []error {
	errCh := make(chan error, len(responses))
	errSlice := make([]error, 0)
	wg := sync.WaitGroup{}
	maxWorkers := runtime.NumCPU()

	if len(responses) < maxWorkers {
		maxWorkers = len(responses)
	}

	responseCh := make(chan *Response, len(responses))
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for response := range responseCh {
				err := r.AddResponse(response)
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	// Send work to workers
	for _, response := range responses {
		responseCh <- response
	}

	close(responseCh)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		errSlice = append(errSlice, err)
	}

	if len(errSlice) > 0 {
		return errSlice
	}

	return nil
}

// GetResponseCount returns the total number of responses stored in the CompressedResponses map.
func (r *CompressResponsePack) GetResponseCount() int {
	var counter int = 0
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, outValue := range r.CompressedResponses {
		counter += len(outValue)
	}
	return counter
}

// GetResponse takes a URL and retrieves a Response from the CompressedResponses map.
// The function handles both direct lookups and URLs with round suffixes.
// It returns the Response object and an error if the response is not found or if decompression fails.
func (r *CompressResponsePack) GetResponse(url string) ([]*Response, error) {
	// Find compressed data

	r.mu.RLock()
	responses, ok := r.CompressedResponses[url]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("response not found for URL: %s", url)
	}
	var responseSlice []*Response

	for _, value := range responses {
		// Decompress
		response, err := NewResponseFromCompressed(value)
		if err != nil {
			return nil, err
		}
		responseSlice = append(responseSlice, response)
	}
	return responseSlice, nil
}

// BatchGetResponse takes a slice of URLs and retrieves the corresponding
// Response objects from the CompressResponsePack concurrently. It handles
// both direct lookups and URLs with round suffixes. The function returns a
// map of maps, where the outer key is the URL and the inner key is the round
// number and the value is the Response object. The function also returns a
// slice of errors if any of the GetResponse operations fail.
func (r *CompressResponsePack) BatchGetResponse(urls []string) (map[string]map[string]*Response, []error) {
	errCh := make(chan error, len(urls))
	respCh := make(chan []*Response, len(urls))
	errSlice := make([]error, 0)

	wg := sync.WaitGroup{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := r.GetResponse(url)
			if err != nil {
				errCh <- err
			} else if resp != nil {
				respCh <- resp
			}
		}(url)
	}

	wg.Wait()
	close(errCh)
	close(respCh)

	for err := range errCh {
		errSlice = append(errSlice, err)
	}

	if len(errSlice) > 0 {
		return nil, errSlice
	}

	responses := make(map[string]map[string]*Response)

	r.mu.Lock()
	defer r.mu.Unlock()

	for resp := range respCh {
		for index, response := range resp {
			responses[response.Url] = make(map[string]*Response)
			newKey := fmt.Sprintf("round_%d", index+1)
			responses[response.Url][newKey] = response
		}
	}
	return responses, nil
}

// DeleteResponse takes a URL and removes the corresponding Responses from the
// CompressedResponses map. If a response does not exist for a URL, it returns an
// error. The function also resets the CompressedResponses map to an empty map if
// no responses remain after deletion.
func (r *CompressResponsePack) DeleteResponse(url string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.CompressedResponses[url]
	if !ok {
		return fmt.Errorf("response not found for URL: %s", url)
	}
	delete(r.CompressedResponses, url)
	if len(r.CompressedResponses) == 0 {
		r.CompressedResponses = make(map[string]map[string][]byte)
	}
	return nil
}

// BatchDeleteResponse takes a slice of URLs and removes the corresponding
// Responses from the CompressedResponses map concurrently. If a response
// does not exist for a URL, it returns an error. The function returns a
// slice of errors for any failed delete operations.
func (r *CompressResponsePack) BatchDeleteResponse(urls []string) []error {
	errCh := make(chan error, len(urls))
	errSlice := make([]error, 0)
	wg := sync.WaitGroup{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := r.DeleteResponse(url)
			if err != nil {
				errCh <- err
			}
		}(url)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		errSlice = append(errSlice, err)
	}

	if len(errSlice) > 0 {
		return errSlice
	}

	return nil
}

// AddInfo adds a key-value pair to the info map of the CompressResponsePack struct.
func (r *CompressResponsePack) AddInfo(key string, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.MetaInfo[key] = value
}

// AddInfoFromMap adds all key-value pairs from the given map to the info map
// of the CompressResponsePack struct.
func (r *CompressResponsePack) AddInfoFromMap(info map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, value := range info {
		r.MetaInfo[key] = value
	}
}

// Clear resets the CompressedResponses map to an empty sync.Map, effectively clearing
// out all the stored responses. This is useful when you want to clear out all the
// responses after they have been processed.
func (r *CompressResponsePack) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CompressedResponses = map[string]map[string][]byte{}
}
