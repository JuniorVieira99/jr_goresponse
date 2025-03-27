package response_test

import (
	"bytes"
	"jr_response/response"
	"strings"
	"testing"

	"github.com/JuniorVieira99/jr_httpcodes/codes"
)

var (
	fixtureResponse, _ = response.NewResponse(
		"https://example.com",
		"example.com",
		codes.GET,
		codes.OK,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"message":"Hello"}`),
		25,
		[]byte(`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 25
Connection: keep-alive

{"message":"Hello world"}`),
	)

	fixtureResponse2, _ = response.NewResponse(
		"https://example2.com",
		"example2.com",
		codes.POST,
		codes.OK,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"message":"Hello two"}`),
		29,
		[]byte(`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 19
Connection: keep-alive

{"message":"Hello world two"}`),
	)

	fixtureRawResponse = []byte(`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 25
Connection: keep-alive

{"message":"Hello world"}`)

	fixtureUrl = "https://example.com"
)

// Response Tests
// -----------------

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		host        string
		method      codes.Method
		statusCode  codes.StatusCode
		headers     map[string]string
		body        []byte
		bodyLength  uint64
		rawResponse []byte
		wantErr     bool
	}{
		{
			name:       "Valid response",
			url:        "https://example.com",
			host:       "example.com",
			method:     codes.GET,
			statusCode: codes.OK,
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       []byte(`{"message":"Hello"}`),
			bodyLength: 19,
			rawResponse: []byte(`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 19
Connection: keep-alive

{"message":"Hello world"}`),
			wantErr: false,
		},
		{
			name:        "Invalid method",
			url:         "https://example.com",
			host:        "example.com",
			method:      "INVALID",
			statusCode:  codes.OK,
			headers:     nil,
			body:        nil,
			bodyLength:  0,
			rawResponse: []byte(`...`),
			wantErr:     true,
		},
		{
			name:        "Invalid status code",
			url:         "https://example.com",
			host:        "example.com",
			method:      codes.GET,
			statusCode:  9999,
			headers:     nil,
			body:        nil,
			bodyLength:  0,
			rawResponse: []byte(`...`),
			wantErr:     true,
		},
		{
			name:       "Nil headers and body",
			url:        "https://example.com",
			host:       "example.com",
			method:     codes.GET,
			statusCode: codes.OK,
			headers:    nil,
			body:       nil,
			bodyLength: 0,
			rawResponse: []byte(`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 19
Connection: keep-alive

{"message":"Hello world"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := response.NewResponse(
				tt.url,
				tt.host,
				tt.method,
				tt.statusCode,
				tt.headers,
				tt.body,
				tt.bodyLength,
				tt.rawResponse,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if resp.Url != tt.url {
					t.Errorf("Response.Url = %v, want %v", resp.Url, tt.url)
				}
				if resp.Host != tt.host {
					t.Errorf("Response.Host = %v, want %v", resp.Host, tt.host)
				}
				if resp.Method != tt.method {
					t.Errorf("Response.Method = %v, want %v", resp.Method, tt.method)
				}
				if resp.StatusCode != tt.statusCode {
					t.Errorf("Response.StatusCode = %v, want %v", resp.StatusCode, tt.statusCode)
				}
				if resp.BodyLength != tt.bodyLength {
					t.Errorf("Response.BodyLength = %v, want %v", resp.BodyLength, tt.bodyLength)
				}
			}
		})
	}
}

func TestResponseToString(t *testing.T) {

	t.Log("TestResponseToString initialization")

	var str string
	str = fixtureResponse.ToString()

	t.Logf("Response.ToString() = %v\n", str)

	if !strings.Contains(str, "example.com") {
		t.Errorf("Response.ToString() = %v, want %v", str, "example.com")
	}

	if !strings.Contains(str, "GET") {
		t.Errorf("Response.ToString() = %v, want %v", str, "GET")
	}

	if !strings.Contains(str, "Content-Type") {
		t.Errorf("Response.ToString() = %v, want %v", str, "Content-Type")
	}

	if !strings.Contains(str, "application/json") {
		t.Errorf("Response.ToString() = %v, want %v", str, "application/json")
	}

	t.Log("TestResponseToString completed")
}

func TestReadBody(t *testing.T) {

	t.Log("TestReadBody initialization")

	str := fixtureResponse.ReadBody()

	t.Logf("Response.ReadBody() = %v\n", str)

	if !strings.Contains(str, "message") {
		t.Errorf("Response.ReadBody() = %v, want %v", str, "Hello world")
	}

	if !strings.Contains(str, "Hello") {
		t.Errorf("Response.ReadBody() = %v, want %v", str, "Hello world")
	}

	t.Log("TestReadBody completed")
}

func TestReadRawResponse(t *testing.T) {

	t.Log("TestReadRawResponse initialization")

	str := fixtureResponse.ReadRawResponse()

	t.Logf("Response.ReadRawResponse() = %v\n", str)

	if !strings.Contains(str, "HTTP/1.1 200 OK") {
		t.Errorf("Response.ReadRawResponse() = %v, want %v", str, "HTTP/1.1 200 OK")
	}

	if !strings.Contains(str, "message") {
		t.Errorf("Response.ReadRawResponse() = %v, want %v", str, "Hello world")
	}

	if !strings.Contains(str, "Hello") {
		t.Errorf("Response.ReadRawResponse() = %v, want %v", str, "Hello world")
	}

	t.Log("TestReadRawResponse completed")
}

func TestToReadableJSON(t *testing.T) {

	t.Log("TestToReadableJSON initialization")

	str, err := fixtureResponse.ToReadableJSON()
	if err != nil {
		t.Errorf("Response.ToJSON() error = %v", err)
	}

	t.Logf("Response.ToJSON() = %v\n", string(str))

	if !bytes.Contains(str, []byte(`"https://example.com"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "https://example.com")
	}

	if !bytes.Contains(str, []byte(`"example.com"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "example.com")
	}

	if !bytes.Contains(str, []byte(`"GET"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "GET")
	}

	if !bytes.Contains(str, []byte(`"statusCode":200`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "200")
	}

	t.Log("TestToReadableJSON completed")

}

func TestToJSON(t *testing.T) {

	t.Log("TestToJSON initialization")

	str, err := fixtureResponse.ToJSON()

	if err != nil {
		t.Errorf("Response.ToJSON() error = %v", err)
	}

	t.Logf("Response.ToJSON() = %v\n", string(str))

	if !bytes.Contains(str, []byte(`"https://example.com"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "https://example.com")
	}

	if !bytes.Contains(str, []byte(`"example.com"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "example.com")
	}

	if !bytes.Contains(str, []byte(`"GET"`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "GET")
	}

	if !bytes.Contains(str, []byte(`"statusCode":200`)) {
		t.Errorf("Response.ToJSON() = %v, want %v", string(str), "200")
	}

	t.Log("TestToJSON completed")
}

func TestNewResponseFromJSON(t *testing.T) {

	t.Log("TestNewResponseFromJSON initialization")

	str, err := fixtureResponse.ToJSON()

	if err != nil {
		t.Errorf("Response.ToJSON() error = %v", err)
	}

	response, err := response.NewResponseFromJSON(str)

	if err != nil {
		t.Errorf("Response.NewResponseFromJSON() error = %v", err)
	}

	t.Logf("Response.NewResponseFromJSON() = %v\n", response.ToString())

	if response.Url != "https://example.com" {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.Url, "https://example.com")
	}

	if response.Host != "example.com" {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.Host, "example.com")
	}

	if response.Method != codes.GET {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.Method, "GET")
	}

	if response.StatusCode != codes.OK {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.StatusCode, "200")
	}

	if !strings.Contains(response.ReadBody(), "message") {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.ReadBody(), "message")
	}

	if !strings.Contains(response.ReadBody(), "Hello") {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.ReadBody(), "Hello")
	}

	if !strings.Contains(response.ReadRawResponse(), "HTTP/1.1 200 OK") {
		t.Errorf("Response.NewResponseFromJSON() = %v, want %v", response.ReadRawResponse(), "HTTP/1.1 200 OK")
	}

	t.Log("TestNewResponseFromJSON completed")

}

func TestNewResponseFromConfig(t *testing.T) {
	config := response.ConfigResponse{
		Method:      codes.GET,
		StatusCode:  codes.OK,
		Url:         "https://example.com",
		Host:        "example.com",
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        []byte(`{"message":"Hello"}`),
		BodyLength:  19,
		RawResponse: []byte(`HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{"message":"Hello"}`),
	}

	response, err := response.NewResponseFromConfig(config)

	if err != nil {
		t.Errorf("Response.NewResponseFromConfig() error = %v", err)
	}

	t.Logf("Response.NewResponseFromConfig() = %v\n", response.ToString())

	if response.Url != "https://example.com" {
		t.Errorf("Response.NewResponseFromConfig() = %v, want %v", response.Url, "https://example.com")
	}

	if response.Host != "example.com" {
		t.Errorf("Response.NewResponseFromConfig() = %v, want %v", response.Host, "example.com")
	}

	if response.Method != codes.GET {
		t.Errorf("Response.NewResponseFromConfig() = %v, want %v", response.Method, "GET")
	}

	if response.StatusCode != codes.OK {
		t.Errorf("Response.NewResponseFromConfig() = %v, want %v", response.StatusCode, "200")
	}

	if !strings.Contains(response.ReadBody(), "message") {
		t.Errorf("Response.NewResponseFromConfig() = %v, want %v", response.ReadBody(), "message")
	}

	t.Log("TestNewResponseFromConfig completed")
}

// Parser
// ------------

func TestResponseParser(t *testing.T) {

	t.Log("TestResponseParser initialization")

	// Create a sample HTTP response
	rawHTTPResponse := []byte(`HTTP/1.1 200 OK
Content-Type: application/json
Server: TestServer/1.0
Date: Sun, 01 Jan 2023 12:00:00 GMT
Content-Length: 27

{"message":"Test successful"}`)

	resp, err := response.ParseRawHTTPResponse(&rawHTTPResponse, fixtureUrl)
	if err != nil {
		t.Fatalf("Failed to parse raw HTTP response: %v", err)
	}

	// Verify parsed response
	if resp.StatusCode != codes.OK {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}

	if resp.Url != fixtureUrl {
		t.Errorf("Expected URL %s, got %s", fixtureUrl, resp.Url)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header application/json, got %s", resp.Headers["Content-Type"])
	}

	if resp.Headers["Server"] != "TestServer/1.0" {
		t.Errorf("Expected Server header TestServer/1.0, got %s", resp.Headers["Server"])
	}

	if !strings.Contains(resp.ReadBody(), "Test successful") {
		t.Errorf("Expected body to contain 'Test successful', got %s", resp.ReadBody())
	}

	// Also test the string version
	strResp, err := response.ParseStringHTTPResponse(string(rawHTTPResponse), fixtureUrl)
	if err != nil {
		t.Fatalf("Failed to parse string HTTP response: %v", err)
	}

	if strResp.StatusCode != codes.OK {
		t.Errorf("String version: Expected status code 200, got %v", strResp.StatusCode)
	}

	t.Log("TestResponseParser completed")
}

// Compression and Decompression
// ------------

func TestCompressAndDecompress(t *testing.T) {
	// Compress the response
	compressed, err := fixtureResponse.Compress()
	if err != nil {
		t.Errorf("Failed to compress response: %v", err)
		return
	}

	// Decompress and verify
	decompressed, err := response.NewResponseFromCompressed(compressed)
	if err != nil {
		t.Errorf("Failed to decompress response: %v", err)
		return
	}

	// Check the decompressed response matches the original
	if decompressed.Url != fixtureResponse.Url {
		t.Errorf("URL mismatch: got %v, want %v", decompressed.Url, fixtureResponse.Url)
	}

	if decompressed.Method != fixtureResponse.Method {
		t.Errorf("Method mismatch: got %v, want %v", decompressed.Method, fixtureResponse.Method)
	}

	if decompressed.StatusCode != fixtureResponse.StatusCode {
		t.Errorf("StatusCode mismatch: got %v, want %v", decompressed.StatusCode, fixtureResponse.StatusCode)
	}

	if !bytes.Equal(decompressed.Body, fixtureResponse.Body) {
		t.Errorf("Body mismatch: got %v, want %v", string(decompressed.Body), string(fixtureResponse.Body))
	}
}

// Edge Cases
// ------------

func TestNewResponseFromJSONWithEmptyData(t *testing.T) {
	emptyData := []byte{}
	_, err := response.NewResponseFromJSON(emptyData)
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message about empty data, got: %v", err)
	}
}

func TestNewResponseFromJSONWithNilData(t *testing.T) {
	var nilData []byte
	_, err := response.NewResponseFromJSON(nilData)
	if err == nil {
		t.Error("Expected error for nil data, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Expected error message about nil data, got: %v", err)
	}
}

func TestNewResponseFromJSONWithInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"url": "https://example.com", "host": "example.com", "method": "GET"`)
	_, err := response.NewResponseFromJSON(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("Expected error about JSON unmarshaling, got: %v", err)
	}
}

func TestParseRawHTTPResponseWithEmptyData(t *testing.T) {
	emptyData := []byte{}
	_, err := response.ParseRawHTTPResponse(&emptyData, "https://example.com")
	if err == nil {
		t.Error("Expected error for empty HTTP response data, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message about empty data, got: %v", err)
	}
}

func TestParseStringHTTPResponseWithEmptyString(t *testing.T) {
	_, err := response.ParseStringHTTPResponse("", "https://example.com")
	if err == nil {
		t.Error("Expected error for empty HTTP response string, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error message about empty data, got: %v", err)
	}
}

func TestParseRawHTTPResponseWithInvalidData(t *testing.T) {
	invalidData := []byte("This is not an HTTP response")
	_, err := response.ParseRawHTTPResponse(&invalidData, "https://example.com")
	if err == nil {
		t.Error("Expected error for invalid HTTP response data, got nil")
	}
}
