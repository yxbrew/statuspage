package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestHealthEndpointInContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container integration test in short mode")
	}

	ctx := context.Background()

	containerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "..",
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"PORT":      "8080",
			"LOG_LEVEL": "info",
		},
		WaitingFor: wait.ForHTTP("/api/v1/health").
			WithPort("8080/tcp").
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}).
			WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start backend container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to resolve container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to resolve mapped port: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/v1/health", host, port.Port())
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	var payload struct {
		Status  string `json:"status"`
		Service string `json:"service"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode health payload: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("unexpected status value: got %q want %q", payload.Status, "ok")
	}

	if payload.Service != "statuspage-backend" {
		t.Fatalf("unexpected service value: got %q want %q", payload.Service, "statuspage-backend")
	}

	if payload.Version == "" {
		t.Fatalf("expected version to be non-empty")
	}
}
