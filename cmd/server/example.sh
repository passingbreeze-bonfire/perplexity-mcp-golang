#!/bin/bash

# Example script to demonstrate running the Perplexity MCP Server
# This script shows the proper way to configure and run the server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Perplexity MCP Server Example${NC}"
echo "========================================"

# Check if API key is provided
if [ -z "$PERPLEXITY_API_KEY" ]; then
    echo -e "${RED}❌ Error: PERPLEXITY_API_KEY environment variable is required${NC}"
    echo ""
    echo "Please set your Perplexity API key:"
    echo "  export PERPLEXITY_API_KEY=\"your-api-key-here\""
    echo ""
    echo "Or run this script with the API key:"
    echo "  PERPLEXITY_API_KEY=\"your-api-key-here\" ./example.sh"
    exit 1
fi

echo -e "${GREEN}✅ API key is configured${NC}"

# Show configuration
echo ""
echo -e "${YELLOW}📋 Current Configuration:${NC}"
echo "  API Key: ${PERPLEXITY_API_KEY:0:8}..." # Show only first 8 characters
echo "  Default Model: ${PERPLEXITY_DEFAULT_MODEL:-llama-3.1-sonar-small-128k-online}"
echo "  Request Timeout: ${REQUEST_TIMEOUT_SECONDS:-30} seconds"
echo "  Log Level: ${LOG_LEVEL:-info}"

# Build the server if it doesn't exist
if [ ! -f "./server" ]; then
    echo ""
    echo -e "${YELLOW}🔨 Building server...${NC}"
    go build -o server cmd/server/main.go
    echo -e "${GREEN}✅ Server built successfully${NC}"
fi

# Run tests first
echo ""
echo -e "${YELLOW}🧪 Running tests...${NC}"
go test ./cmd/server -v

echo ""
echo -e "${GREEN}✅ All tests passed!${NC}"

# Start the server with a timeout for demonstration
echo ""
echo -e "${YELLOW}🖥️  Starting server for 5 seconds...${NC}"
echo "  (In production, remove the timeout to run indefinitely)"

# Run the server in background and capture PID
./server &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Show server is running
if kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${GREEN}✅ Server is running (PID: $SERVER_PID)${NC}"
    echo ""
    echo "Available MCP tools:"
    echo "  🔍 perplexity_search   - Search for information"
    echo "  💬 perplexity_chat    - Chat with AI assistant"
    echo "  📚 perplexity_research - Comprehensive research"
else
    echo -e "${RED}❌ Server failed to start${NC}"
    exit 1
fi

# Wait a bit more
sleep 3

# Gracefully shutdown
echo ""
echo -e "${YELLOW}⏹️  Sending shutdown signal...${NC}"
kill -TERM $SERVER_PID

# Wait for graceful shutdown
wait $SERVER_PID 2>/dev/null || true

echo -e "${GREEN}✅ Server shut down gracefully${NC}"
echo ""
echo -e "${GREEN}🎉 Example completed successfully!${NC}"
echo ""
echo "To run the server in production:"
echo "  PERPLEXITY_API_KEY=\"your-key\" ./server"