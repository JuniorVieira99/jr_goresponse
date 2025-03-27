# Compressed Response Pack

## Overview

The compressed response pack is a data structure that is used to represent a pack of HTTP responses in a compressed format.

## Index

- [Overview](#overview)
- [Index](#index)
- [Why?](#why)
- [Constructor](#constructor)
  - [NewCompressResponsePack](#newcompressresponsepack)
- [Structure](#structure)
- [Methods](#methods)
  - [AddResponse](#addresponse)
  - [BatchAddResponse](#batchaddresponse)
  - [GetResponse](#getresponse)
  - [BatchGetResponse](#batchgetresponse)
  - [GetResponseCount](#getresponsecount)
  - [DeleteResponse](#deleteresponse)
  - [BatchDeleteResponse](#batchdeleteresponse)
  - [AddInfo](#addinfo)
  - [AddInfoFromMap](#addinfofrommap)
  - [Clear](#clear)
- [Tests](#tests)
- [Usage Example](#usage-example)

## Why?

When dealing with a large number of responses, it can be beneficial to compress the response data to save space and reduce memory usage. The compressed response pack provides a way to store and retrieve compressed response data efficiently.

## Constructor

### NewCompressResponsePack

```go
func NewCompressResponsePack() *CompressResponsePack
```

## Structure

| Field | Type | Description |
| --- | --- | --- |
| CompressedResponses | map[string]map[string][]byte | Map storing compressed Response data where outer keys are URLs, inner keys are rounds, and values are gzip-compressed byte arrays |
| MetaInfo | map[string]string | Map storing additional metadata about the response pack |
| mu | sync.RWMutex | Mutex for ensuring thread-safe operations |

## Methods

### AddResponse

```go
func (r *CompressResponsePack) AddResponse(response *Response) error
```

Compresses the given Response object and adds it to the CompressedResponses map, handling duplicate URLs by appending a round suffix.

### BatchAddResponse

```go
func (r *CompressResponsePack) BatchAddResponse(responses []*Response) []error
```

Compresses and adds multiple Response objects concurrently to the CompressResponsePack, returning any errors encountered.

### GetResponse

```go
func (r *CompressResponsePack) GetResponse(url string) ([]*Response, error)
```

Retrieves and decompresses Response objects for a given URL, returning them as a slice.

### BatchGetResponse

```go
func (r *CompressResponsePack) BatchGetResponse(urls []string) (map[string]map[string]*Response, []error)
```

Concurrently retrieves and decompresses Response objects for multiple URLs, returning a map of results and any errors.

### GetResponseCount

```go
func (r *CompressResponsePack) GetResponseCount() int
```

Returns the total number of compressed responses stored in the pack.

### DeleteResponse

```go
func (r *CompressResponsePack) DeleteResponse(url string) error
```

Removes all Response objects for a given URL from the CompressedResponses map.

### BatchDeleteResponse

```go
func (r *CompressResponsePack) BatchDeleteResponse(urls []string) []error
```

Concurrently removes Response objects for multiple URLs, returning any errors encountered.

### AddInfo

```go
func (r *CompressResponsePack) AddInfo(key string, value string) 
```

Adds a key-value pair to the MetaInfo map of the CompressResponsePack.

### AddInfoFromMap

```go
func (r *CompressResponsePack) AddInfoFromMap(info map[string]string)
```

Adds all key-value pairs from the given map to the MetaInfo map.

### Clear

```go
func (r *CompressResponsePack) Clear()
```

Resets the CompressedResponses map to empty, clearing all stored responses.

## Tests

To run the tests, execute the following command from the root of the repository:

```bash
go test ./tests/response_compress_pack_test.go
```

## Usage Example

```go
// Create a new CompressResponsePack
cpack := response.NewCompressResponsePack()

// Create a response
resp, _ := response.NewResponse(
    "https://example.com",
    "example.com",
    codes.GET,
    codes.OK,
    nil,
    []byte("Hello World"),
    11,
    nil,
)

// Add the response to the pack -> will automatically compress it
cpack.AddResponse(resp)

// Get the response back (will be decompressed)
responses, _ := cpack.GetResponse("https://example.com")
for _, r := range responses {
    fmt.Println("Response body:", string(r.Body))
}

// Add metadata
cpack.AddInfo("created", time.Now().String())

// Get total count
fmt.Println("Total responses:", cpack.GetResponseCount())
```
