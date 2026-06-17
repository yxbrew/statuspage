package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestHealthEndpointInContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping container integration test in short mode")
	}

	ctx := context.Background()
	networkName := fmt.Sprintf("statuspage-it-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

	network, err := tcnetwork.New(
		ctx,
		tcnetwork.WithDriver("bridge"),
		tcnetwork.WithLabels(map[string]string{"name": networkName}),
	)
	if err != nil {
		t.Fatalf("failed to create integration network: %v", err)
	}
	networkName = network.Name
	t.Cleanup(func() {
		_ = network.Remove(ctx)
	})

	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "statuspage",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"statuspage-postgres"},
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(90 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = postgresContainer.Terminate(context.Background())
	})

	imageTag := fmt.Sprintf("statuspage-backend:it-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())
	_, testFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve integration test path")
	}
	backendDir := filepath.Dir(filepath.Dir(testFilePath))

	buildCmd := exec.Command("docker", "build", "-t", imageTag, "-f", "Dockerfile", ".")
	buildCmd.Dir = backendDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build backend image: %v, output: %s", err, string(output))
	}

	containerReq := testcontainers.ContainerRequest{
		Image:        imageTag,
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"PORT":         "8080",
			"LOG_LEVEL":    "info",
			"DATABASE_URL": "postgres://postgres:postgres@statuspage-postgres:5432/statuspage?sslmode=disable",
		},
		Networks: []string{networkName},
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
