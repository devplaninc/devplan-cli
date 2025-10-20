package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RunServer() error {
	server := mcp.NewServer(&mcp.Implementation{Name: "devplan", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "reportWorkLog", Description: "reportWorkLog"}, reportWorkLog)
	return server.Run(context.Background(), &mcp.StdioTransport{})
}
