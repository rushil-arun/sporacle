package gameinit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var forwardClient = &http.Client{Timeout: 5 * time.Second}

// ForwardCreate sends a CreateRequest to targetAddr/internal/create-game
// and returns the parsed CreateResponse. Used when the chosen server is not self.
func ForwardCreate(ctx context.Context, targetAddr string, req CreateRequest) (CreateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://"+targetAddr+"/internal/create-game",
		bytes.NewReader(body),
	)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := forwardClient.Do(httpReq)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("forward request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CreateResponse{}, fmt.Errorf("target server returned %d", resp.StatusCode)
	}

	var result CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return CreateResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
