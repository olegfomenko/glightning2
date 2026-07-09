package clnrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

const jsonrpcVersion = "2.0"

// Client calls Core Lightning's native JSON-RPC interface over lightning-rpc.
type Client struct {
	socketPath string
	nextID     atomic.Uint64
}

// NewClient returns a JSON-RPC client for a Core Lightning lightning-rpc Unix
// socket path.
func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

// RPCError is an error returned by Core Lightning's JSON-RPC server.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return fmt.Sprintf("cln rpc error %d", e.Code)
	}
	return fmt.Sprintf("cln rpc error %d: %s", e.Code, e.Message)
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	ID      string          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// Call sends one JSON-RPC request to Core Lightning and decodes result into
// result. If result is nil, the response result is ignored.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if c == nil {
		return errors.New("clnrpc: nil client")
	}
	if c.socketPath == "" {
		return errors.New("clnrpc: empty socket path")
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Time{})
	}

	id := fmt.Sprintf("%d", c.nextID.Add(1))
	req := rpcRequest{
		JSONRPC: jsonrpcVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp rpcResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	if result == nil {
		return nil
	}
	if len(resp.Result) == 0 {
		return errors.New("clnrpc: response missing result")
	}
	return json.Unmarshal(resp.Result, result)
}
