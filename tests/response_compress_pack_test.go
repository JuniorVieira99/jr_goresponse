package response_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/JuniorVieira99/jr_goresponse/response"
	"github.com/JuniorVieira99/jr_httpcodes/codes"
)

// Test fixtures
var (
	compressTestResp1, _ = response.NewResponse(
		"https://example.com/api1",
		"example.com",
		codes.GET,
		codes.OK,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"result":"success1"}`),
		0,
		[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"result\":\"success1\"}"),
	)

	compressTestResp2, _ = response.NewResponse(
		"https://example.com/api2",
		"example.com",
		codes.POST,
		codes.Created,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"result":"success2"}`),
		0,
		[]byte("HTTP/1.1 201 Created\r\nContent-Type: application/json\r\n\r\n{\"result\":\"success2\"}"),
	)

	compressTestResp3, _ = response.NewResponse(
		"https://example.com/api3",
		"example.com",
		codes.GET,
		codes.NotFound,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"error":"not found"}`),
		0,
		[]byte("HTTP/1.1 404 Not Found\r\nContent-Type: application/json\r\n\r\n{\"error\":\"not found\"}"),
	)
)

func TestNewCompressResponsePack(t *testing.T) {
	pack := response.NewCompressResponsePack()

	if pack == nil {
		t.Fatal("NewCompressResponsePack() returned nil")
	}

	if pack.CompressedResponses == nil {
		t.Error("CompressResponsePack.CompressedResponses is nil")
	}

	if pack.MetaInfo == nil {
		t.Error("CompressResponsePack.MetaInfo is nil")
	}
}

func TestCompressResponseAddResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()
	if pack == nil {
		t.Fatal("NewCompressResponsePack() returned nil")
	}

	err := pack.AddResponse(compressTestResp1)
	if err != nil {
		t.Fatalf("AddResponse() error = %v", err)
	}

	// Check if response was added
	count := pack.GetResponseCount()
	if count != 1 {
		t.Errorf("After AddResponse(), count = %d, want 1", count)
	}

	// Adding nil response should return error
	err = pack.AddResponse(nil)
	if err == nil {
		t.Error("AddResponse(nil) should return error")
	}
}

func TestCompressResponseMap(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)
	_ = pack.AddResponse(compressTestResp2)

	// Test mapping
	mapping := pack.CompressedResponses

	// Verify mapping
	if _, ok := mapping["https://example.com/api1"]; !ok {
		t.Error("Map() missing response for https://example.com/api1")
	}

	if _, ok := mapping["https://example.com/api2"]; !ok {
		t.Error("Map() missing response for https://example.com/api2")
	}
}

func TestCompressResponseGetResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)

	// Test getting existing response
	resp, err := pack.GetResponse("https://example.com/api1")
	if err != nil {
		t.Fatalf("GetResponse() error = %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("GetResponse() returned %d responses, want 1", len(resp))
	}

	if resp[0].Url != "https://example.com/api1" {
		t.Errorf("GetResponse().Url = %s, want https://example.com/api1", resp[0].Url)
	}

	// Test getting non-existent response
	_, err = pack.GetResponse("https://nonexistent.com")
	if err == nil {
		t.Error("GetResponse() for non-existent URL should return error")
	}

	// Test multiple rounds for same URL
	_ = pack.AddResponse(compressTestResp1) // Add again to create a second round
	resp, err = pack.GetResponse("https://example.com/api1")
	if err != nil {
		t.Fatalf("GetResponse() error = %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("GetResponse() for URL with multiple rounds returned %d responses, want 2", len(resp))
	}

}

func TestCompressResponseBatchAddResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Create batch of responses
	responses := []*response.Response{compressTestResp1, compressTestResp2, compressTestResp3}
	errs := pack.BatchAddResponse(responses)

	if errs != nil {
		t.Fatalf("BatchAddResponse() returned errors: %v", errs)
	}

	if pack.GetResponseCount() != 3 {
		t.Errorf("After BatchAddResponse(), count = %d, want 3", pack.GetResponseCount())
	}

	// Test batch with some nil responses
	nilBatch := []*response.Response{nil, compressTestResp1}
	errs = pack.BatchAddResponse(nilBatch)

	if len(errs) == 0 {
		t.Error("BatchAddResponse() with nil response should return errors")
	}
}

func TestCompressResponseBatchGetResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)
	_ = pack.AddResponse(compressTestResp3)

	// Test batch retrieval
	urls := []string{"https://example.com/api1", "https://example.com/api2"}
	results, errs := pack.BatchGetResponse(urls)

	if errs != nil {
		t.Fatalf("BatchGetResponse() returned errors: %v", errs)
	}

	if len(results) != 2 {
		t.Errorf("BatchGetResponse() returned %d results, want 2", len(results))
	}

	// Should have entries for both URLs
	if _, ok := results["https://example.com/api1"]; !ok {
		t.Error("BatchGetResponse() missing result for https://example.com/api1")
	}

	if _, ok := results["https://example.com/api2"]; !ok {
		t.Error("BatchGetResponse() missing result for https://example.com/api2")
	}

	// Test with non-existent URL
	badUrls := []string{"https://example.com/api1", "https://nonexistent.com"}
	_, errs = pack.BatchGetResponse(badUrls)

	if len(errs) == 0 {
		t.Error("BatchGetResponse() with non-existent URL should return errors")
	}
}

func TestCompressResponseDeleteResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)

	// Delete existing response
	err := pack.DeleteResponse("https://example.com/api1")
	if err != nil {
		t.Fatalf("DeleteResponse() error = %v", err)
	}

	// Verify deletion
	_, err = pack.GetResponse("https://example.com/api1")
	if err == nil {
		t.Error("GetResponse() should return error after DeleteResponse()")
	}

	// Delete non-existent response
	err = pack.DeleteResponse("https://nonexistent.com")
	if err == nil {
		t.Error("DeleteResponse() for non-existent URL should return error")
	}
}

func TestCompressResponseBatchDeleteResponse(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)
	_ = pack.AddResponse(compressTestResp3)

	// Batch delete
	errs := pack.BatchDeleteResponse([]string{"https://example.com/api1", "https://example.com/api2"})
	if len(errs) > 0 {
		t.Fatalf("BatchDeleteResponse() returned errors: %v", errs)
	}

	// Verify deletion
	_, err := pack.GetResponse("https://example.com/api1")
	if err == nil {
		t.Error("GetResponse() should return error after BatchDeleteResponse()")
	}

	// Check remaining response
	count := pack.GetResponseCount()
	if count != 1 {
		t.Errorf("After BatchDeleteResponse(), count = %d, want 1", count)
	}

	// Batch delete with non-existent URL
	errs = pack.BatchDeleteResponse([]string{"https://example.com/api3", "https://nonexistent.com"})
	if len(errs) != 1 {
		t.Errorf("BatchDeleteResponse() with non-existent URL returned %d errors, want 1", len(errs))
	}
}

func TestCompressResponseAddInfo(t *testing.T) {
	pack := response.NewCompressResponsePack()

	pack.AddInfo("test_key", "test_value")

	if value, ok := pack.MetaInfo["test_key"]; !ok || value != "test_value" {
		t.Errorf("AddInfo() didn't set MetaInfo[\"test_key\"] correctly, got %v", value)
	}

	pack.AddInfo("test_key", "updated_value")

	if value, ok := pack.MetaInfo["test_key"]; !ok || value != "updated_value" {
		t.Errorf("AddInfo() didn't update MetaInfo[\"test_key\"] correctly, got %v", value)
	}
}

func TestCompressResponseAddInfoFromMap(t *testing.T) {
	pack := response.NewCompressResponsePack()

	infoMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	pack.AddInfoFromMap(infoMap)

	if value, ok := pack.MetaInfo["key1"]; !ok || value != "value1" {
		t.Errorf("AddInfoFromMap() didn't set MetaInfo[\"key1\"] correctly, got %v", value)
	}

	if value, ok := pack.MetaInfo["key2"]; !ok || value != "value2" {
		t.Errorf("AddInfoFromMap() didn't set MetaInfo[\"key2\"] correctly, got %v", value)
	}
}

func TestCompressResponseClear(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Add responses
	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)

	// Clear the pack
	pack.Clear()

	// Verify it's empty
	count := pack.GetResponseCount()
	if count != 0 {
		t.Errorf("After Clear(), count = %d, want 0", count)
	}
}

func TestCompressResponseGetResponseCount(t *testing.T) {
	pack := response.NewCompressResponsePack()

	if pack.GetResponseCount() != 0 {
		t.Errorf("Empty CompressResponsePack GetResponseCount() = %d, want 0", pack.GetResponseCount())
	}

	_ = pack.AddResponse(compressTestResp1)
	_ = pack.AddResponse(compressTestResp2)

	if pack.GetResponseCount() != 2 {
		t.Errorf("After adding 2 responses, GetResponseCount() = %d, want 2", pack.GetResponseCount())
	}
}

func TestConcurrentAccessCompressPack(t *testing.T) {
	pack := response.NewCompressResponsePack()

	// Test concurrent access to CompressResponsePack
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(3) // 3 goroutines

	// Goroutine 1: Add responses
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = pack.AddResponse(compressTestResp1)
		}
	}()

	// Goroutine 2: Get responses
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_, _ = pack.GetResponse("https://example.com/api1")
		}
	}()

	// Goroutine 3: Add info
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			pack.AddInfo("test_key", "test_value")
		}
	}()

	wg.Wait()

	// No assertion needed; if there's a race condition, the race detector will catch it
}

func TestCompressResponseCompressionEffectiveness(t *testing.T) {
	// Create a large response with repetitive content to test compression effectiveness
	largeBody := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog.", 1000))

	largeResp, _ := response.NewResponse(
		"https://example.com/large",
		"example.com",
		codes.GET,
		codes.OK,
		map[string]string{"Content-Type": "text/plain"},
		largeBody,
		uint64(len(largeBody)),
		nil,
	)

	// Compress the response directly to verify compression works
	compressed, err := largeResp.Compress()
	if err != nil {
		t.Fatalf("Failed to compress response: %v", err)
	}

	// Check that compression is effective (compressed size should be smaller)
	if len(compressed) >= len(largeBody) {
		t.Errorf("Compression not effective: original %d bytes, compressed %d bytes",
			len(largeBody), len(compressed))
	}

	// Now test the round-trip through CompressResponsePack
	pack := response.NewCompressResponsePack()

	err = pack.AddResponse(largeResp)
	if err != nil {
		t.Fatalf("AddResponse() error = %v", err)
	}

	// Retrieve the response
	retrievedResps, err := pack.GetResponse("https://example.com/large")
	if err != nil {
		t.Fatalf("GetResponse() error = %v", err)
	}

	// Verify the response was properly decompressed
	if len(retrievedResps) != 1 {
		t.Fatalf("GetResponse() returned %d responses, want 1", len(retrievedResps))
	}

	if len(retrievedResps[0].Body) != len(largeBody) {
		t.Errorf("Decompressed body length %d, want %d",
			len(retrievedResps[0].Body), len(largeBody))
	}

	if string(retrievedResps[0].Body) != string(largeBody) {
		t.Error("Decompressed body content doesn't match original")
	}
}
