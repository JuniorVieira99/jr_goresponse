// Package response provides utilities for handling HTTP responses in a structured way.
//
// The package offers types and functions to create, manipulate, and analyze HTTP response
// data. It includes the core Response struct for individual HTTP responses and ResponsePack
// for managing collections of responses with thread-safe operations.
//
// The package supports:
// - Creating responses from raw HTTP data or structured configurations
// - Converting responses to and from JSON format
// - Thread-safe storage and retrieval of multiple responses
// - Statistics calculation for response success/failure rates
// - Parsing raw HTTP response data
//
// Core types:
// - Response: Represents a single HTTP response with method, status, headers, and body
// - ResponsePack: Thread-safe collection of responses with aggregated statistics
// - ConfigResponse: Configuration structure for creating new Response instances
//
// This package is designed to work with the codes package for HTTP method and status code
// validation and handling.
package response

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlPack "net/url"
	"request_comp/codes"
	"strings"
	"sync"
)

// Response struct
// ----------------------------------------------------------------------

// Response struct
type Response struct {
	Method     codes.Method      `json:"method"`
	StatusCode codes.StatusCode  `json:"statusCode"`
	Url        string            `json:"url"`
	Host       string            `json:"host"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	BodyLength uint64            `json:"bodyLength"`
}

type ConfigResponse struct {
	Method     codes.Method
	StatusCode codes.StatusCode
	Url        string
	Host       string
	Headers    map[string]string
	Body       []byte
	BodyLength uint64
}

// String returns a string representation of the Response struct, including
// the URL, host, method, status code, headers, body, and body length formatted
// in a readable manner.
func (r *Response) ToString() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Url: %s", r.Url))
	sb.WriteString(fmt.Sprintf("\nHost: %s", r.Host))
	sb.WriteString(fmt.Sprintf("\nMethod: %s", r.Method))
	sb.WriteString(fmt.Sprintf("\nStatusCode: %d", r.StatusCode))

	sb.WriteString("\nHeaders:")
	for key, value := range r.Headers {
		sb.WriteString(fmt.Sprintf("\n%s: %s", key, value))
	}

	// Write each byte directly to prevent truncation issues
	sb.WriteString("\nBody: ")
	if len(r.Body) > 0 {
		sb.Write(r.Body) // This properly handles all bytes
	} else {
		sb.WriteString("<empty>")
	}
	sb.WriteString(fmt.Sprintf("\nBodyLength: %d", r.BodyLength))

	return sb.String()
}

// Print prints a string representation of the Response struct to the console.
func (r *Response) Print() {
	fmt.Println(r.ToString())
}

// ToJSON converts the Response struct into a JSON-encoded byte slice.
// It returns an error if the encoding process fails.
func (r *Response) ToJSON() ([]byte, error) {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// Constructors
// ----------------------------------------------------------------------

// NewResponseFromJSON takes a JSON-encoded byte slice and attempts to decode it
// into a Response struct. It returns a pointer to the Response struct and a possible
// error if the decoding fails.
func NewResponseFromJSON(data []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
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

	return &Response{
		Method:     method,
		StatusCode: statusCode,
		Url:        url,
		Host:       host,
		Headers:    headers,
		Body:       body,
		BodyLength: bodyLength,
	}, nil
}

// NewResponseFromConfig creates a new Response instance using the provided ResponseConfig.
// It returns a pointer to the Response struct and a possible error if the Response cannot
// be created. This function leverages the NewResponse function to perform validation and
// initialization of the Response fields.
func NewResponseFromConfig(config ConfigResponse) (*Response, error) {
	return NewResponse(config.Url, config.Host, config.Method, config.StatusCode, config.Headers, config.Body, config.BodyLength)
}

// Response Pack
// ----------------------------------------------------------------------

// ResponsePack A struct with many Responses objects
type ResponsePack struct {
	Responses    sync.Map `json:"responses"`
	Total        uint64   `json:"total"`
	Success      uint64   `json:"success"`
	Failure      uint64   `json:"failure"`
	SuccessRatio float64  `json:"successRatio"`
	FailureRatio float64  `json:"failureRatio"`
	Info         sync.Map `json:"info"`
	Mu           sync.RWMutex
}

func (r *ResponsePack) GetResponse(url string) *Response {
	// Try direct lookup first
	value, ok := r.Responses.Load(url)
	if ok {
		response, _ := value.(*Response)
		return response
	}

	// Try looking for URL with round suffix
	var result *Response
	r.Responses.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true // continue
		}

		// Check if this key is the URL with a round suffix
		if strings.HasPrefix(keyStr, url+"--round_") || keyStr == url {
			result = value.(*Response)
			return false // stop iteration
		}
		return true // continue
	})

	return result
}

// GetIndexes takes a URL and returns a slice of all the indexes where the URL can be found in the ResponsePack.
func (p *ResponsePack) GetIndexes(url string) []int {
	var counter = 0
	var sliceIndex []int

	p.Mu.RLock()
	defer p.Mu.RUnlock()

	p.Responses.Range(func(key, value any) bool {
		validKey, ok := key.(string)
		if !ok {
			return true
		}

		if validKey == "" {
			return true
		}

		// Extract the base URL
		baseURL := validKey
		if roundIndex := strings.Index(validKey, "--round_"); roundIndex >= 0 {
			baseURL = validKey[:roundIndex]
		}

		// Compare with the requested URL
		if baseURL == url {
			sliceIndex = append(sliceIndex, counter)
		}

		counter++
		return true
	})
	return sliceIndex
}

// GetKeysOfResponses returns a slice containing all the keys present in the Responses map of the ResponsePack.
// It acquires a read lock on the mutex to ensure thread-safe access to the map.
func (p *ResponsePack) GetKeysOfResponses() []string {

	p.Mu.RLock()
	defer p.Mu.RUnlock()

	var keys []string

	p.Responses.Range(func(key, value any) bool {
		keys = append(keys, key.(string))
		return true
	})

	return keys
}

// AddResponse adds a Response to the ResponsePack struct. It will create a new
// inner map if the URL doesn't exist yet, or use the existing map if it does.
// The Response is stored with the index as the key in the inner map.
func (p *ResponsePack) AddResponse(response *Response) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	ok := codes.IsSuccess(response.StatusCode)
	if ok {
		p.Success++
	} else {
		p.Failure++
	}
	p.Total++

	// Recalculate ratios directly after updating metrics
	p.Calculate()

	// Handle duplicate URLs
	origURL := response.Url
	urlKey := origURL

	// Find a unique key
	round := 0
	for {
		_, exists := p.Responses.Load(urlKey)
		if !exists {
			break
		}
		round++
		urlKey = fmt.Sprintf("%s--round_%d", origURL, round)
	}

	if urlKey != origURL {
		responseCopy := *response // Create a copy
		responseCopy.Url = urlKey // Update the copy's URL
		p.Responses.Store(urlKey, &responseCopy)
	} else {
		p.Responses.Store(urlKey, response)
	}

	return nil
}

// Calculate recalculates the success and failure ratios of the ResponsePack.
func (p *ResponsePack) Calculate() {
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
	p.Info.Store(key, value)
}

// ToString returns a string representation of the ResponsePack struct,
// including the total number of responses, number of successful responses,
// number of failed responses, and the success and failure ratios. It also includes the info map as key-value pairs.
func (p *ResponsePack) ToString() string {

	var str strings.Builder

	str.WriteString(fmt.Sprintf("Total: %d", p.Total))
	str.WriteString(fmt.Sprintf("\nSuccess: %d", p.Success))
	str.WriteString(fmt.Sprintf("\nFailure: %d", p.Failure))
	str.WriteString(fmt.Sprintf("\nSuccessRatio: %f", p.SuccessRatio))
	str.WriteString(fmt.Sprintf("\nFailureRatio: %f", p.FailureRatio))
	str.WriteString(fmt.Sprint("\nInfo:"))

	p.Info.Range(func(key, value interface{}) bool {
		str.WriteString(fmt.Sprintf("\n%v: %v", key, value))
		return true
	})

	return str.String()
}

// Print prints a string representation of the ResponsePack struct to the console.
func (p *ResponsePack) Print() {
	fmt.Println(p.ToString())
}

// Len returns the number of responses stored in the ResponsePack.
func (p *ResponsePack) Len() int {
	var counter int = 0
	p.Responses.Range(func(key, value any) bool {
		counter++
		return true
	})
	return counter
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
func (p *ResponsePack) GetErrorReport() (map[string]string, error) {

	if p.Len() == 0 {
		return nil, fmt.Errorf("no responses found")
	}

	if p == nil {
		return nil, fmt.Errorf("response pack is nil")
	}

	p.Mu.RLock()
	defer p.Mu.RUnlock()

	output := make(map[string]string, p.Len())

	p.Responses.Range(func(key, value any) bool {
		resp := value.(*Response)

		if !codes.IsSuccess(resp.StatusCode) {
			output[key.(string)] = resp.StatusCode.String()
		}

		return true
	})
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
		str.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}

	return str.String(), nil
}

// NewResponsePack returns a new ResponsePack instance with zero values for all fields.
func NewResponsePack() *ResponsePack {
	return &ResponsePack{
		Responses:    sync.Map{}, // keys: string, values: *Response
		Info:         sync.Map{}, // keys: string, values: string
		Total:        0,
		Success:      0,
		Failure:      0,
		SuccessRatio: 0,
		FailureRatio: 0,
		Mu:           sync.RWMutex{},
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
