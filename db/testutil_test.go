package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal("start postgres container:", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal("open db:", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := conn.Ping(); err != nil {
		t.Fatal("ping:", err)
	}

	schema, err := os.ReadFile(filepath.Join(".", "schema.sql"))
	if err != nil {
		t.Fatal("read schema.sql:", err)
	}
	if _, err := conn.ExecContext(ctx, string(schema)); err != nil {
		t.Fatal("exec schema:", err)
	}

	return conn
}
