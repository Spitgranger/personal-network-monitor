package checks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

type HTTPResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	DurationMS int64  `json:"duration_ms"`
	Status     string `json:"status"`
	ErrorCode  string `json:"error_code,omitempty"`
}

func CheckHttp(ctx context.Context, client *http.Client, target string) HTTPResult {
	result := HTTPResult{URL: target, Status: "failure"}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		result.ErrorCode = "Invalid Request"
		return result
	}
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			result.ErrorCode = "Request cancelled by user"

		case errors.Is(err, context.DeadlineExceeded):
			result.ErrorCode = "Request timed out"

		default:
			result.ErrorCode = "HTTP request failed:"
		}
		duration := time.Since(startTime)
		result.DurationMS = duration.Milliseconds()
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode
	_, err = io.ReadAll(resp.Body)
	duration := time.Since(startTime)
	result.DurationMS = duration.Milliseconds()
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			result.ErrorCode = "Request cancelled by user"

		case errors.Is(err, context.DeadlineExceeded):
			result.ErrorCode = "Request timed out"
		default:
			result.ErrorCode = "Failed to read request body"
		}
		return result
	}
	// define success as 200 OK
	if resp.StatusCode == http.StatusOK {
		result.Status = "success"
	}
	result.ErrorCode = ""
	return result
}
