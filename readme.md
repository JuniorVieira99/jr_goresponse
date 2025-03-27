# Response Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/JuniorVieira99/response_lib.svg)](https://pkg.go.dev/github.com/JuniorVieira99/response_lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/JuniorVieira99/response_lib)](https://goreportcard.com/report/github.com/JuniorVieira99/response_lib)

A comprehensive Go library for handling HTTP responses in a structured and thread-safe way. This library provides utilities for creating, parsing, manipulating, and analyzing HTTP response data.

## Overview

This library have three main structs:

- `Response`: Represents an HTTP response with metadata such as URL, host, method, status code, headers, body, body length, and the raw response.
- `ResponsePack`: Represents a collection of `Response` objects, with support for concurrent access and metadata.
- `CompressResponsePack`: Represents a collection of compressed `Response` objects, with support for concurrent access and metadata.

## Features

- **Structured Response Handling**: Represent HTTP responses with comprehensive metadata
- **Thread-safe Response Collections**: Manage multiple responses with concurrent access support
- **HTTP Response Parsing**: Parse raw HTTP response data into structured objects
- **Statistics Generation**: Calculate success/failure rates and other metrics
- **JSON Serialization**: Convert responses to and from JSON format
- **Flexible Creation Options**: Create responses via direct instantiation or configuration objects
- **Error Reporting**: Generate detailed error reports for failed requests
- **Metadata Support**: Attach custom metadata to response packs
- **Docs**: Check docs directory for detailed documentation

## Installation

```bash
go get github.com/JuniorVieira99/response_lib
```

## Dependencies

This library depends on:

```bash
go get github.com/JuniorVieira99/jr_httpcodes
```

## Usage

### Creating a Response

```go
// *NOTE*: For more detailed explanation of methods check docs

import (
    "github.com/JuniorVieira99/jr_httpcodes/codes"
    "github.com/JuniorVieira99/response_lib/response"
)

headers := map[string]string{"Content-Type": "application/json"}
body := []byte(`{"message":"Hello"}`)

// Create a response directly
resp, err := response.NewResponse(
    "https://example.com",       // URL
    "example.com",               // Host
    codes.GET,                   // Method
    codes.OK,                    // Status code
    headers,                     // Headers
    body,                        // Body
    19,                          // Body length
    []byte(`HTTP/1.1 200 OK....`), // Raw response
)

// Create a response from configuration
config := response.ConfigResponse{
    Url:        "https://example.com",
    Host:       "example.com",
    Method:     codes.GET,
    StatusCode: codes.OK,
    Headers:    map[string]string{"Content-Type": "application/json"},
    Body:       []byte(`{"message":"Hello"}`),
    BodyLength: 19,
    RawResponse: []byte(`HTTP/1.1 200 OK....`),
}
resp, err := response.NewResponseFromConfig(config)
if err != nil {
    // Handle error
}
```

### Parsing Raw HTTP Responses

```go
rawData := []byte(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 19

{"message":"Hello"}`)

// This will create a response struct directly from a raw response
resp, err := response.ParseRawHTTPResponse(&rawData, "https://example.com")
if err != nil {
    // Handle error
}

// You can also parse from string
strData := `HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 19

{"message":"Hello"}`

resp, err = response.ParseStringHTTPResponse(strData, "https://example.com")
if err != nil {
    // Handle error
}
```

### Working with Response Collections

```go

// *NOTE*: For more detailed explanation of methods check docs

// Create a response pack
pack := response.NewResponsePack()

// Add responses
pack.AddResponse(resp1)
pack.AddResponse(resp2)

// Could also responses in batches
responses := []*response.Response{resp1, resp2}
pack.BatchAddResponse(responses)

// Calculate statistics, then use ToString() or Print() to display
pack.Calculate()

// Get a specific response
resp := pack.GetResponse("https://example.com")

// Get all keys, which are the URLs of the responses
keys := pack.GetKeysOfResponses()

// Get error report
errorReport, err := pack.GetErrorReport()
if err != nil {
    // Handle error
}

// Get error report as string
errorReportStr, err := pack.GetErrorReportString()
if err != nil {
    // Handle error
}

// Add metadata
pack.AddInfo("requestTime", "2023-01-01T12:00:00Z")
pack.AddInfo("requestSource", "API Test")
```

### Converting to JSON

```go
// Convert response to JSON
jsonData, err := resp.ToJSON()
if err != nil {
    // Handle error
}

// Create response from JSON
newResp, err := response.NewResponseFromJSON(jsonData)
if err != nil {
    // Handle error
}
```

### String Representation

```go
// Get string representation of a response
respStr := resp.ToString()
fmt.Println(respStr)

// Or print directly
resp.Print()

// Get string representation of a response pack
packStr := pack.ToString()
fmt.Println(packStr)

// Or print directly
pack.Print()
```

### Compressing Response Pack

```go
// *NOTE*: For more detailed explanation of methods check docs

// Create a compress response pack
compressPack := response.NewCompressResponsePack()

// Add responses, the responses will be automatically compressed
compressPack.AddResponse(resp1)
compressPack.AddResponse(resp2)
compressPack.AddResponse(resp3)

// You could also add responses in batch
responses := []*response.Response{resp1, resp2, resp3}
compressPack.BatchAddResponse(responses)

// Get a specific response, the response will be automatically decompressed
resp := compressPack.GetResponse("https://example.com")

// Also, allows for batch get
urls := []string{"https://example.com/api1", "https://example.com/api2"}
responses := compressPack.BatchGetResponse(urls)
```

## Thread Safety

The `ResponsePack` type is designed to be thread-safe, using a combination of `sync.Map` for storing responses and metadata, and a `sync.RWMutex` for protecting other fields. This makes it suitable for concurrent operations in multi-goroutine environments.

## Error Handling

The library provides comprehensive error handling:

- All constructor functions return errors when validation fails
- `GetErrorReport()` method generates a map of failed responses
- `GetErrorReportString()` provides a human-readable error summary

## Examples

### Complete Example: Parsing and Analyzing HTTP Responses

```go
package main

import (
    "fmt"
    "github.com/JuniorVieira99/response_lib/response"
)

func main() {
    // Create a response pack
    pack := response.NewResponsePack()
    
    // Parse some raw HTTP responses
    rawResp1 := []byte(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 19

{"message":"Hello"}`)
    
    rawResp2 := []byte(`HTTP/1.1 404 Not Found
Content-Type: text/plain
Content-Length: 9

Not Found`)
    
    // Parse and add responses
    resp1, _ := response.ParseRawHTTPResponse(&rawResp1, "https://example.com/api1")
    resp2, _ := response.ParseRawHTTPResponse(&rawResp2, "https://example.com/api2")
    
    pack.AddResponse(resp1)
    pack.AddResponse(resp2)
    
    // Add metadata
    pack.AddInfo("testSuite", "API Integration Tests")
    pack.AddInfo("runDate", "2023-01-15")
    
    // Calculate statistics
    pack.Calculate()
    
    // Print summary
    fmt.Println("Response Pack Summary:")
    pack.Print()
    
    // Get error report
    errorReport, _ := pack.GetErrorReportString()
    fmt.Println("\nError Report:")
    fmt.Println(errorReport)
}
```

## Complete Example: Parsing and Compressing HTTP Responses

```go
package main

import (
    "fmt"
    "github.com/JuniorVieira99/response_lib/response"
)

func main() {
    // Create a response pack
    pack := response.NewResponsePack()
    
    // Parse some raw HTTP responses
    rawResp1 := []byte(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 19

{"message":"Hello"}`)
    
    rawResp2 := []byte(`HTTP/1.1 404 Not Found
Content-Type: text/plain
Content-Length: 9

Not Found`)
    
    // Parse and add responses
    resp1, _ := response.ParseRawHTTPResponse(&rawResp1, "https://example.com/api1")
    resp2, _ := response.ParseRawHTTPResponse(&rawResp2, "https://example.com/api2")

    // Create a compress response pack
    compressPack := response.NewCompressResponsePack()

    // Add responses
    compressPack.AddResponse(resp1)
    compressPack.AddResponse(resp2)

    // Get a specific response
    resp := compressPack.GetResponse("https://example.com/api1")

    // Check response
    resp.IsSuccessful()
}
```

## Tests

To run the tests for the `Response` struct, execute the following command from the root of the repository:

```bash
go test ./tests/response_test.go
```

To run the tests for the `ResponsePack` struct, execute the following command from the root of the repository:

```bash
go test ./tests/response_pack_test.go
```

To run the tests for the `Compress Response Pack` structs, execute the following command from the root of the repository:

```bash
go test ./tests/response_compress_pack_test.go
```

To test race conditions, execute the following command from the root of the repository:

```bash
go test -race ./...
```

To specific directory or test file use:

```bash
go test -race ./tests/response_test.go
```

## Docs

Check the `docs` directory in the repository for detailed documentation.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
