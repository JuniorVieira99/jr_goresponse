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
	testResp1, _ = response.NewResponse(
		"https://example.com/api1",
		"example.com",
		codes.GET,
		codes.OK,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"result":"success1"}`),
		0,
		[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"result\":\"success1\"}"),
	)

	testResp2, _ = response.NewResponse(
		"https://example.com/api2",
		"example.com",
		codes.POST,
		codes.Created,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"result":"success2"}`),
		0,
		[]byte("HTTP/1.1 201 Created\r\nContent-Type: application/json\r\n\r\n{\"result\":\"success2\"}"),
	)

	testResp3, _ = response.NewResponse(
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

func TestNewResponsePack(t *testing.T) {
	pack := response.NewResponsePack()

	if pack == nil {
		t.Fatal("NewResponsePack() returned nil")
	}

	if pack.Responses == nil {
		t.Error("ResponsePack.Responses is nil")
	}

	if pack.Info == nil {
		t.Error("ResponsePack.Info is nil")
	}

	if pack.Total != 0 {
		t.Errorf("ResponsePack.Total = %d, want 0", pack.Total)
	}

	if pack.Success != 0 {
		t.Errorf("ResponsePack.Success = %d, want 0", pack.Success)
	}

	if pack.Failure != 0 {
		t.Errorf("ResponsePack.Failure = %d, want 0", pack.Failure)
	}
}

func TestAddResponse(t *testing.T) {
	pack := response.NewResponsePack()

	err := pack.AddResponse(testResp1)
	if err != nil {
		t.Fatalf("AddResponse() error = %v", err)
	}

	// Check metrics
	if pack.Total != 1 {
		t.Errorf("After AddResponse(), Total = %d, want 1", pack.Total)
	}

	if pack.Success != 1 {
		t.Errorf("After AddResponse() with OK status, Success = %d, want 1", pack.Success)
	}

	// Add failure response
	err = pack.AddResponse(testResp3)
	if err != nil {
		t.Fatalf("AddResponse() error = %v", err)
	}

	if pack.Total != 2 {
		t.Errorf("After second AddResponse(), Total = %d, want 2", pack.Total)
	}

	if pack.Failure != 1 {
		t.Errorf("After AddResponse() with 404 status, Failure = %d, want 1", pack.Failure)
	}

	// Check ratios
	if pack.SuccessRatio != 0.5 {
		t.Errorf("SuccessRatio = %f, want 0.5", pack.SuccessRatio)
	}

	if pack.FailureRatio != 0.5 {
		t.Errorf("FailureRatio = %f, want 0.5", pack.FailureRatio)
	}

	// Adding nil response should return error
	err = pack.AddResponse(nil)
	if err == nil {
		t.Error("AddResponse(nil) should return error")
	}
}

func TestGetResponse(t *testing.T) {
	pack := response.NewResponsePack()

	// Add responses
	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp2)

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
	_ = pack.AddResponse(testResp1) // Add again to create a second round
	resp, err = pack.GetResponse("https://example.com/api1")
	if err != nil {
		t.Fatalf("GetResponse() error = %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("GetResponse() for URL with multiple rounds returned %d responses, want 2", len(resp))
	}
}

func TestBatchGetResponse(t *testing.T) {
	pack := response.NewResponsePack()

	// Add responses
	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp2)
	_ = pack.AddResponse(testResp3)

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

func TestGetKeysOfResponses(t *testing.T) {
	pack := response.NewResponsePack()

	// Add responses
	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp2)

	keys := pack.GetKeysOfResponses()

	if len(keys) != 2 {
		t.Errorf("GetKeysOfResponses() returned %d keys, want 2", len(keys))
	}

	// Check if both URLs are in the keys
	foundApi1 := false
	foundApi2 := false

	for _, key := range keys {
		if key == "https://example.com/api1" {
			foundApi1 = true
		} else if key == "https://example.com/api2" {
			foundApi2 = true
		}
	}

	if !foundApi1 {
		t.Error("GetKeysOfResponses() missing key https://example.com/api1")
	}

	if !foundApi2 {
		t.Error("GetKeysOfResponses() missing key https://example.com/api2")
	}
}

func TestBatchAddResponse(t *testing.T) {
	pack := response.NewResponsePack()

	// Create batch of responses
	responses := []*response.Response{testResp1, testResp2, testResp3}
	errs := pack.BatchAddResponse(responses)

	if errs != nil {
		t.Fatalf("BatchAddResponse() returned errors: %v", errs)
	}

	if pack.Total != 3 {
		t.Errorf("After BatchAddResponse(), Total = %d, want 3", pack.Total)
	}

	if pack.Success != 2 {
		t.Errorf("After BatchAddResponse(), Success = %d, want 2", pack.Success)
	}

	if pack.Failure != 1 {
		t.Errorf("After BatchAddResponse(), Failure = %d, want 1", pack.Failure)
	}

	// Test batch with some nil responses
	nilBatch := []*response.Response{nil, testResp1}
	errs = pack.BatchAddResponse(nilBatch)

	if errs == nil || len(errs) == 0 {
		t.Error("BatchAddResponse() with nil response should return errors")
	}
}

func TestCalculate(t *testing.T) {
	pack := response.NewResponsePack()

	// Empty pack should not cause division by zero
	pack.Calculate()

	// Add responses
	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp3)

	// Calculate manually
	pack.Success = 1
	pack.Failure = 1
	pack.Total = 2
	pack.Calculate()

	if pack.SuccessRatio != 0.5 {
		t.Errorf("Calculate() set SuccessRatio = %f, want 0.5", pack.SuccessRatio)
	}

	if pack.FailureRatio != 0.5 {
		t.Errorf("Calculate() set FailureRatio = %f, want 0.5", pack.FailureRatio)
	}
}

func TestAddInfo(t *testing.T) {
	pack := response.NewResponsePack()

	pack.AddInfo("test_key", "test_value")

	if value, ok := pack.Info["test_key"]; !ok || value != "test_value" {
		t.Errorf("AddInfo() didn't set Info[\"test_key\"] correctly, got %v", value)
	}

	pack.AddInfo("test_key", "updated_value")

	if value, ok := pack.Info["test_key"]; !ok || value != "updated_value" {
		t.Errorf("AddInfo() didn't update Info[\"test_key\"] correctly, got %v", value)
	}
}

func TestToString(t *testing.T) {
	pack := response.NewResponsePack()

	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp2)
	_ = pack.AddResponse(testResp3)

	pack.AddInfo("test_key", "test_value")

	str := pack.ToString()

	if !strings.Contains(str, "Total: 3") {
		t.Errorf("ToString() = %v, want %v", str, "Total: 3")
	}

	if !strings.Contains(str, "Success: 2") {
		t.Errorf("ToString() = %v, want %v", str, "Success: 2")
	}

	if !strings.Contains(str, "Failure: 1") {
		t.Errorf("ToString() = %v, want %v", str, "Failure: 1")
	}

	if !strings.Contains(str, "SuccessRatio: 0.666667") {
		t.Errorf("ToString() = %v, want %v", str, "SuccessRatio: 0.666667")
	}

	if !strings.Contains(str, "FailureRatio: 0.333333") {
		t.Errorf("ToString() = %v, want %v", str, "FailureRatio: 0.333333")
	}

	if !strings.Contains(str, "test_key: test_value") {
		t.Errorf("ToString() = %v, want %v", str, "test_key: test_value")
	}
}

func TestGetErrorReport(t *testing.T) {
	pack := response.NewResponsePack()

	// Add responses with different status codes
	_ = pack.AddResponse(testResp1) // 200 OK
	_ = pack.AddResponse(testResp2) // 201 Created
	_ = pack.AddResponse(testResp3) // 404 Not Found

	errorReport, err := pack.GetErrorReport()
	if err != nil {
		t.Fatalf("GetErrorReport() error = %v", err)
	}

	// Should have only the failed response
	if len(errorReport) != 1 {
		t.Errorf("GetErrorReport() returned %d URLs, want 1", len(errorReport))
	}

	if _, ok := errorReport["https://example.com/api3"]; !ok {
		t.Error("GetErrorReport() missing error for https://example.com/api3")
	}

	// Test empty pack
	emptyPack := response.NewResponsePack()
	_, err = emptyPack.GetErrorReport()
	if err == nil {
		t.Error("GetErrorReport() on empty pack should return error")
	}
}

func TestGetErrorReportString(t *testing.T) {
	pack := response.NewResponsePack()

	// Add responses with different status codes
	_ = pack.AddResponse(testResp1) // 200 OK
	_ = pack.AddResponse(testResp3) // 404 Not Found

	report, err := pack.GetErrorReportString()
	if err != nil {
		t.Fatalf("GetErrorReportString() error = %v", err)
	}

	t.Logf("GetErrorReportString() = %v", report)

	expectedContents := []string{
		"Error Report:",
		"https://example.com/api3",
		"404",
	}

	if !strings.Contains(report, expectedContents[0]) {
		t.Errorf("GetErrorReportString() = %v, want %v", report, expectedContents[0])
	}

	if !strings.Contains(report, expectedContents[1]) {
		t.Errorf("GetErrorReportString() = %v, want %v", report, expectedContents[1])
	}

	if !strings.Contains(report, expectedContents[2]) {
		t.Errorf("GetErrorReportString() = %v, want %v", report, expectedContents[2])
	}

	// Test empty pack
	emptyPack := response.NewResponsePack()
	_, err = emptyPack.GetErrorReportString()
	if err == nil {
		t.Error("GetErrorReportString() on empty pack should return error")
	}
}

func TestConcurrentAccess(t *testing.T) {
	pack := response.NewResponsePack()

	// Test concurrent access to ResponsePack
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(3) // 3 goroutines

	// Goroutine 1: Add responses
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = pack.AddResponse(testResp1)
		}
	}()

	// Goroutine 2: Get responses
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_, _ = pack.GetResponse("https://example.com/api1")
		}
	}()

	// Goroutine 3: Get keys
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = pack.GetKeysOfResponses()
		}
	}()

	wg.Wait()

	// No assertion needed; if there's a race condition, the race detector will catch it
}

func TestLen(t *testing.T) {
	pack := response.NewResponsePack()

	if pack.Len() != 0 {
		t.Errorf("Empty ResponsePack Len() = %d, want 0", pack.Len())
	}

	_ = pack.AddResponse(testResp1)
	_ = pack.AddResponse(testResp2)

	if pack.Len() != 2 {
		t.Errorf("After adding 2 responses, Len() = %d, want 2", pack.Len())
	}

	// Test nil pack
	var nilPack *response.ResponsePack
	if nilPack.Len() != 0 {
		t.Errorf("Nil ResponsePack Len() = %d, want 0", nilPack.Len())
	}
}
