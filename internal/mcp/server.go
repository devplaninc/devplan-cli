package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/specsync"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer() *Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "devplan", Version: "v1.0.0"}, nil)
	server := &Server{srv: srv, syncers: make(map[string]*specsync.Syncer)}
	mcp.AddTool(srv, &mcp.Tool{Name: "reportWorkLog", Description: "reportWorkLog"}, server.reportWorkLog)
	return server
}

type Server struct {
	srv     *mcp.Server
	syncers map[string]*specsync.Syncer

	mu sync.Mutex
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("MCP Server: starting")

	defer func() {
		if r := recover(); r != nil {
			slog.Error("MCP server: panicked", "panic", r)
			return
		}
		slog.Info("MCP server: stopped")
	}()
	return s.srv.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) addSyncer(ctx context.Context, companyID int32, taskID string) error {
	if companyID <= 0 || taskID == "" {
		return fmt.Errorf("invalid parameters: companyID=%d, taskID=%s", companyID, taskID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d-%s", companyID, taskID)
	if _, ok := s.syncers[key]; ok {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	client := devplan.NewClient(devplan.Config{})
	adapter := specsync.NewClientAdapter(client)
	specsResp, err := client.GetTaskSpecs(companyID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task specs: %w", err)
	}

	taskDir := specsResp.GetPathsInfo().GetTaskPaths()[taskID].GetTaskDir()
	if taskDir == "" {
		return nil
	}
	fullTaskDir := filepath.Join(cwd, taskDir)
	interval := specsync.DefaultSyncInterval
	syncer := specsync.NewSyncer(adapter, companyID, taskID, fullTaskDir, interval)
	s.syncers[key] = syncer
	go syncer.RunBackground(ctx)
	go syncer.TriggerOnce(ctx)
	slog.Info("Started syncer", "companyID", companyID, "taskID", taskID)
	return nil
}
