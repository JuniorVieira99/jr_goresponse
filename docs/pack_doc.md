# Response Pack

## Overview

The response pack is a data structure that is used to represent a pack of HTTP responses.

## Index

- [Overview](#overview)
- [Index](#index)
- [Why?](#why)
- [Structure](#structure)
- [Methods](#methods)
  - [AddResponse](#addresponse)
  - [GetResponse](#getresponse)
  - [Calculate](#calculate)
  - [GetErrorReport](#geterrorreport)
  - [GetIndexes](#getindexes)
  - [GetKeysOfResponses](#getkeysofresponses)
  - [ToString](#tostring)
  - [Print](#print)
- [Tests](#tests)
- [Usage Example](#usage-example)

## Why?

The response pack is useful for storing and managing multiple responses in a single data structure. It provides methods for adding responses, retrieving responses by URL, calculating success and failure ratios, and generating error reports.

## Structure

| Field | Type | Description |
| --- | --- | --- |
| Responses | sync.Map | Thread-safe map storing Response objects where keys are URLs and values are Response pointers |
| Total | uint64 | Total number of responses in the pack |
| Success | uint64 | Number of successful responses (2xx status codes) |
| Failure | uint64 | Number of failed responses (non-2xx status codes) |
| SuccessRatio | float64 | Ratio of successful responses to total responses |
| FailureRatio | float64 | Ratio of failed responses to total responses |
| Info | sync.Map | Thread-safe map storing additional metadata about the response pack |
| Mu | sync.RWMutex | Mutex for ensuring thread-safe operations |

## Methods

### AddResponse

```go
func (p *ResponsePack) AddResponse(response *Response) error
```

Adds a Response to the ResponsePack, updating statistics and handling duplicate URLs by appending a round suffix.

### GetResponse

```go
func (r *ResponsePack) GetResponse(url string) *Response
```

Retrieves a Response by URL, handling both direct lookups and URLs with round suffixes.

### Calculate

```go
func (p *ResponsePack) Calculate()
```

Recalculates the success and failure ratios based on current counts.

### GetErrorReport

```go
func (p *ResponsePack) GetErrorReport() (map[string]string, error)
```

Returns a map of URLs to status codes for all failed responses.

### GetIndexes

```go
func (p *ResponsePack) GetIndexes(url string) []int
```

Returns all indexes where a specific URL can be found in the ResponsePack.

### GetKeysOfResponses

```go
func (p *ResponsePack) GetKeysOfResponses() []string
```

Returns a slice containing all keys present in the Responses map.

### ToString

```go
func (p *ResponsePack) ToString() string
```

Returns a formatted string representation of the ResponsePack including statistics and info.

### Print

```go
func (p *ResponsePack) Print()
```

Prints the string representation of the ResponsePack to the console.

## Tests

To run the tests, execute the following command from the root of the repository:

```bash
go test ./tests/response_pack_test.go
```

## Usage Example

```go
// Create a new ResponsePack
pack := response.NewResponsePack()

// Create a response
resp, _ := response.NewResponse(
    "https://example.com",
    "example.com",
    codes.GET,
    codes.OK,
    nil,
    []byte("Hello World"),
    11,
)

// Add the response to the pack
pack.AddResponse(resp)

// Get statistics
fmt.Println("Total responses:", pack.Total)
fmt.Println("Success ratio:", pack.SuccessRatio)

// Get error report
errorReport, _ := pack.GetErrorReportString()
fmt.Println(errorReport)
```
