# Response

## Overview

The response object is a struct representing the response from the server. It contains the status code, headers, and body of the response.

## Index

- [Overview](#overview)
- [Index](#index)
- [Why?](#why)
- [Structure](#structure)
- [Methods](#methods)
  - [ToString](#tostring)
  - [ReadBody](#readbody)
  - [ReadRawResponse](#readrawresponse)
  - [Print](#print)
  - [ToReadableJSON](#toreadablejson)
  - [ToJSON](#tojson)
  - [Compress](#compress)
- [Constructors](#constructors)
  - [NewResponseFromJSON](#newresponsefromjson)
  - [NewResponse](#newresponse)
  - [NewResponseFromConfig](#newresponsefromconfig)
  - [NewResponseFromCompressed](#newresponsefromcompressed)
- [Parser Functions](#parser-functions)
  - [ParseRawHTTPResponse](#parserawhttpresponse)
  - [ParseStringHTTPResponse](#parsestringhttpresponse)
- [Tests](#tests)
- [Usage Example](#usage-example)

## Why?

The response object is useful for storing and managing the response data from an HTTP request. It provides methods for converting the response to a string, printing the response to the console, converting the response to JSON, and compression and decompression.

## Structure

| Field | Type | Description |
| --- | --- | --- |
| Method | codes.Method | The HTTP method used in the request. |
| StatusCode | codes.StatusCode | The status code of the response. |
| Url | string | The URL of the request. |
| Host | string | The host of the request. |
| Headers | map[string]string | The headers of the response. |
| Body | []byte | The body of the response. |
| Body Length | uint64 | The length of the body. |
| RawResponse | []byte | The raw response data. |

## Methods

### ToString

```go
func (r *Response) ToString() string
```

Returns a string representation of the Response object, including URL, host, method, status code, headers, body, and body length.

### ReadBody

```go
func (r *Response) ReadBody() string
```

Returns the response body as a string.

### ReadRawResponse

```go
func (r *Response) ReadRawResponse() string
```

Returns the raw response data as a string.

### Print

```go
func (r *Response) Print()
```

Prints a string representation of the Response struct to the console.

### ToReadableJSON

```go
func (r *Response) ToReadableJSON() ([]byte, error)
```

Converts the Response struct to a JSON-encoded byte slice, with special handling for binary data for human readability.

**Example**:

```go
// Create a new Response
someResponse, _ = response.NewResponse(
    "https://example.com",
    "example.com",
    codes.GET,
    codes.OK,
    map[string]string{"Content-Type": "application/json"},
    []byte(`{"message":"Hello"}`),
    25,
    []byte(
`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 25
Connection: keep-alive

{"message":"Hello world"}`
),
)

// Convert to JSON
jsonData, _ := someResponse.ToReadableJSON()

// Print the JSON data
fmt.Println(string(jsonData))
```

**Output**:

```json
{"method":"GET","statusCode":200,"url":"https://example.com","host":"example.com","headers":{"Content-Type":"application/json"},"body":"{\"message\":\"Hello\"}","bodyLength":25,"rawResponse":"HTTP/1.1 200 OK\nServer: nginx/1.18.0\nDate: Mon, 01 Jan 2023 12:00:00 GMT\nContent-Type: application/json\nContent-Length: 25\nConnection: keep-alive\n\n{\"message\":\"Hello world\"}","encoding":{}}
```

### ToJSON

```go
func (r *Response) ToJSON() ([]byte, error)
```

Converts the Response struct to a JSON-encoded byte slice.

**Example**:

```go
// Create a new Response
someResponse, _ = response.NewResponse(
    "https://example.com",
    "example.com",
    codes.GET,
    codes.OK,
    map[string]string{"Content-Type": "application/json"},
    []byte(`{"message":"Hello"}`),
    25,
    []byte(
`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 25
Connection: keep-alive

{"message":"Hello world"}`
),
)

// Convert to JSON
jsonData, _ := someResponse.ToJSON()

// Print the JSON data
fmt.Println(string(jsonData))
```

**Output**:

```json
{"method":"GET","statusCode":200,"url":"https://example.com","host":"example.com","headers":{"Content-Type":"application/json"},"body":"eyJtZXNzYWdlIjoiSGVsbG8ifQ==","bodyLength":25,"rawResponse":"SFRUUC8xLjEgMjAwIE9LClNlcnZlcjogbmdpbngvMS4xOC4wCkRhdGU6IE1vbiwgMDEgSmFuIDIwMjMgMTI6MDA6MDAgR01UCkNvbnRlbnQtVHlwZTogYXBwbGljYXRpb24vanNvbgpDb250ZW50LUxlbmd0aDogMjUKQ29ubmVjdGlvbjoga2VlcC1hbGl2ZQoKeyJtZXNzYWdlIjoiSGVsbG8gd29ybGQifQ=="}
```

### Compress

```go
func (r *Response) Compress() ([]byte, error)
```

Compresses the Response object using Gzip compression.

## Constructors

### NewResponseFromJSON

```go
func NewResponseFromJSON(data []byte) (*Response, error)
```

Creates a new Response from a JSON-encoded byte slice.

### NewResponse

```go
func NewResponse(url string, host string, method codes.Method, statusCode codes.StatusCode, headers map[string]string, body []byte, bodyLength uint64, rawResponse []byte) (*Response, error)
```

Creates a new Response instance with the given parameters.

### NewResponseFromConfig

```go
func NewResponseFromConfig(config ConfigResponse) (*Response, error)
```

Creates a new Response instance using the provided ResponseConfig.

### NewResponseFromCompressed

```go
func NewResponseFromCompressed(compressedData []byte) (*Response, error)
```

Creates a Response from compressed data.

## Parser Functions

### ParseRawHTTPResponse

```go
func ParseRawHTTPResponse(rawResponse *[]byte, url string) (*Response, error)
```

Parses raw HTTP response data into a Response struct.

### ParseStringHTTPResponse

```go
func ParseStringHTTPResponse(rawResponse string, url string) (*Response, error)
```

Parses HTTP response data from a string into a Response struct.

## Tests

To run the tests, execute the following command from the root of the repository:

```bash
go test ./tests/response_test.go
```

## Usage Example

```go
// Create a new Response
resp, err := response.NewResponse(
    "https://example.com",           
    "example.com",
    codes.GET,
    codes.OK,
    map[string]string{"Content-Type": "application/json"},
    []byte(`{"message": "Hello World"}`),
    19,
    nil,
)

// Get the response body
body := resp.ReadBody()

// Get the raw response
rawResponse := resp.ReadRawResponse()

// Convert to a string
respStr := resp.ToString()

// Print the response
resp.Print()

// Convert to JSON
jsonData, err := resp.ToJSON()

// Parse a raw HTTP response
RawResponse = []byte(
`HTTP/1.1 200 OK
Server: nginx/1.18.0
Date: Mon, 01 Jan 2023 12:00:00 GMT
Content-Type: application/json
Content-Length: 25
Connection: keep-alive

{"message":"Hello world"}`
)

parsedResp, err := response.ParseRawHTTPResponse(&RawResponse, "https://example.com")
```