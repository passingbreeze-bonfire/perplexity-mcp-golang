package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper structures for JSON-RPC communication
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// TestHelper manages subprocess communication for integration tests
type TestHelper struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	t      *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	// Set required environment variable for testing
	os.Setenv("PERPLEXITY_API_KEY", "test-key-for-integration-tests")

	cmd := exec.Command("go", "run", "./main.go")

	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	stderr, err := cmd.StderrPipe()
	require.NoError(t, err)

	return &TestHelper{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		t:      t,
	}
}

func (th *TestHelper) Start() {
	require.NoError(th.t, th.cmd.Start())
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
}

func (th *TestHelper) Stop() {
	if th.cmd.Process != nil {
		// Gracefully close stdin first
		th.stdin.Close()

		// Wait for process to finish, with timeout
		done := make(chan error, 1)
		go func() {
			done <- th.cmd.Wait()
		}()

		select {
		case <-done:
			// Process finished naturally
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't finish
			th.cmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			th.cmd.Process.Kill()
		}
	}

	// Close remaining pipes
	if th.stdout != nil {
		th.stdout.Close()
	}
	if th.stderr != nil {
		th.stderr.Close()
	}
}

func (th *TestHelper) SendRequest(req JSONRPCRequest) {
	data, err := json.Marshal(req)
	require.NoError(th.t, err)

	_, err = th.stdin.Write(append(data, '\n'))
	require.NoError(th.t, err)
}

func (th *TestHelper) ReadResponse() JSONRPCResponse {
	reader := bufio.NewReader(th.stdout)
	line, _, err := reader.ReadLine()
	require.NoError(th.t, err)

	var resp JSONRPCResponse
	err = json.Unmarshal(line, &resp)
	require.NoError(th.t, err)

	return resp
}

func (th *TestHelper) ReadResponseWithTimeout(timeout time.Duration) (JSONRPCResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	responseChan := make(chan JSONRPCResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(th.stdout)
		line, _, err := reader.ReadLine()
		if err != nil {
			errorChan <- err
			return
		}

		var resp JSONRPCResponse
		err = json.Unmarshal(line, &resp)
		if err != nil {
			errorChan <- err
			return
		}

		responseChan <- resp
	}()

	select {
	case resp := <-responseChan:
		return resp, nil
	case err := <-errorChan:
		return JSONRPCResponse{}, err
	case <-ctx.Done():
		return JSONRPCResponse{}, fmt.Errorf("timeout waiting for response")
	}
}

func (th *TestHelper) ReadStderr() string {
	// Read available stderr content non-blocking
	buffer := make([]byte, 4096)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	var n int
	go func() {
		n, _ = th.stderr.Read(buffer)
		close(done)
	}()

	select {
	case <-done:
		return string(buffer[:n])
	case <-ctx.Done():
		return "" // Timeout, no data available
	}
}

// Test cases

func TestStdioTransportBasic(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Test basic JSON-RPC communication with initialize request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(req)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 1, int(resp.ID.(float64)))
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)
}

func TestStdioTransportToolsList(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// First initialize the server
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	helper.ReadResponseWithTimeout(2 * time.Second)

	// Now test tools/list
	toolsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      2,
	}

	helper.SendRequest(toolsReq)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 2, int(resp.ID.(float64)))
	assert.Nil(t, resp.Error)

	// Verify we have the perplexity_search tool
	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok)
	require.Len(t, tools, 1)

	tool, ok := tools[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "perplexity_search", tool["name"])
}

func TestStdioTransportValidRequest(t *testing.T) {
	// Skip this test if we don't have a real API key
	if os.Getenv("PERPLEXITY_API_KEY") == "" || os.Getenv("PERPLEXITY_API_KEY") == "test-key-for-integration-tests" {
		t.Skip("Skipping test that requires real Perplexity API key")
	}

	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Initialize the server first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	helper.ReadResponseWithTimeout(2 * time.Second)

	// Test valid search request
	searchReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: ToolCallParams{
			Name: "perplexity_search",
			Arguments: map[string]interface{}{
				"query": "What is the capital of France?",
				"model": "sonar",
			},
		},
		ID: 2,
	}

	helper.SendRequest(searchReq)

	resp, err := helper.ReadResponseWithTimeout(30 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 2, int(resp.ID.(float64)))
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)

	// Verify result structure
	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)

	content, ok := result["content"].([]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)

	textContent, ok := content[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "text", textContent["type"])
	assert.NotEmpty(t, textContent["text"])
}

func TestStdioTransportMalformedInput(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Send malformed JSON
	malformedInput := []byte("{invalid json\n")
	_, err := helper.stdin.Write(malformedInput)
	require.NoError(t, err)

	// Should get an error response or the connection should handle gracefully
	// Wait a moment for processing
	time.Sleep(500 * time.Millisecond)

	// Verify stderr contains error logs
	stderrContent := helper.ReadStderr()
	assert.NotEmpty(t, stderrContent)

	// Now send a valid request to ensure server is still responsive
	validReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(validReq)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, "2.0", resp.JSONRPC)
}

func TestStdioTransportEmptyInput(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Send empty line
	_, err := helper.stdin.Write([]byte("\n"))
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Server should handle gracefully and remain responsive
	validReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(validReq)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, "2.0", resp.JSONRPC)
}

func TestStdioTransportLargePayload(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	helper.ReadResponseWithTimeout(2 * time.Second)

	// Create a large query string
	largeQuery := strings.Repeat("This is a test query for large payload handling. ", 100)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: ToolCallParams{
			Name: "perplexity_search",
			Arguments: map[string]interface{}{
				"query": largeQuery,
				"model": "sonar",
			},
		},
		ID: 2,
	}

	helper.SendRequest(req)

	// Should get a response - could be error or success depending on validation
	resp, err := helper.ReadResponseWithTimeout(10 * time.Second)
	if err != nil {
		// Log stderr for debugging
		stderrContent := helper.ReadStderr()
		t.Logf("Stderr content: %s", stderrContent)
		t.Fatalf("Failed to read response: %v", err)
	}

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 2, int(resp.ID.(float64)))
	// Note: Large query might be accepted or rejected depending on validation
}

func TestStdioTransportInvalidToolName(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	helper.ReadResponseWithTimeout(2 * time.Second)

	// Test invalid tool name
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: ToolCallParams{
			Name: "nonexistent_tool",
			Arguments: map[string]interface{}{
				"query": "test",
			},
		},
		ID: 2,
	}

	helper.SendRequest(req)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 2, int(resp.ID.(float64)))
	assert.NotNil(t, resp.Error)
}

func TestStdioTransportInvalidMethodName(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Test invalid method name
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "invalid/method",
		ID:      1,
	}

	helper.SendRequest(req)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 1, int(resp.ID.(float64)))
	assert.NotNil(t, resp.Error)
}

func TestStdioTransportStderrLogsOnly(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Send a request to generate logs
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(req)

	resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	// Verify response is valid JSON-RPC
	assert.Equal(t, "2.0", resp.JSONRPC)

	// Check that stderr contains log messages
	stderrContent := helper.ReadStderr()
	assert.Contains(t, stderrContent, "[MAIN]") // Should contain main process logs

	// Verify that stdout only contains JSON-RPC responses, no log pollution
	// This is implicitly tested by successful JSON parsing of responses
}

func TestStdioTransportResourceCleanup(t *testing.T) {
	helper := NewTestHelper(t)

	helper.Start()

	// Send a simple request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(req)

	// Get response to ensure server is working
	_, err := helper.ReadResponseWithTimeout(5 * time.Second)
	require.NoError(t, err)

	// Stop the server
	helper.Stop()

	// Verify process has exited
	assert.True(t, helper.cmd.ProcessState != nil)
}

func TestStdioTransportSequentialRequests(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Stop()

	helper.Start()

	// Initialize first
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	initResp, err := helper.ReadResponseWithTimeout(2 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, 1, int(initResp.ID.(float64)))

	// Send multiple sequential requests
	for i := 2; i <= 4; i++ {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/list",
			ID:      i,
		}
		helper.SendRequest(req)

		resp, err := helper.ReadResponseWithTimeout(5 * time.Second)
		require.NoError(t, err)

		assert.Equal(t, "2.0", resp.JSONRPC)
		assert.Equal(t, i, int(resp.ID.(float64)))
		assert.Nil(t, resp.Error)
	}
}

// Benchmark test for stdio transport performance
func BenchmarkStdioTransport(b *testing.B) {
	helper := NewTestHelper(&testing.T{})
	defer helper.Stop()

	helper.Start()

	// Initialize
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]interface{}{},
			ClientInfo: ClientInfo{
				Name:    "benchmark-client",
				Version: "1.0.0",
			},
		},
		ID: 1,
	}

	helper.SendRequest(initReq)
	helper.ReadResponseWithTimeout(2 * time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "tools/list",
			ID:      i + 2,
		}

		helper.SendRequest(req)
		helper.ReadResponseWithTimeout(5 * time.Second)
	}
}
