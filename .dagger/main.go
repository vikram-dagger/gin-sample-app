package main

import (
	"context"
	"dagger/book/internal/dagger"
)

type Book struct{}

func (m *Book) Run(
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Service {
	postgresdb := dag.Container().
		From("postgres:alpine").
		WithEnvVariable("POSTGRES_DB", "app_test").
		WithEnvVariable("POSTGRES_PASSWORD", "secret").
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})

	return m.Env(source).
		WithServiceBinding("db", postgresdb).
		WithEnvVariable("DB_HOST", "db").
		WithEnvVariable("DB_USER", "postgres").
		WithEnvVariable("DB_PASSWORD", "secret").
		WithEnvVariable("DB_NAME", "app_test").
		WithEnvVariable("DB_PORT", "5432").
		WithExec([]string{"go", "run", "main.go"}).
		WithExposedPort(8080).
		AsService()
}

func (m *Book) Test(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	postgresdb := dag.Container().
		From("postgres:alpine").
		WithEnvVariable("POSTGRES_DB", "app_test").
		WithEnvVariable("POSTGRES_PASSWORD", "secret").
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})

	return m.Env(source).
		WithServiceBinding("db", postgresdb).
		WithEnvVariable("DB_HOST", "db").
		WithEnvVariable("DB_USER", "postgres").
		WithEnvVariable("DB_PASSWORD", "secret").
		WithEnvVariable("DB_NAME", "app_test").
		WithEnvVariable("DB_PORT", "5432").
		WithExec([]string{"go", "test", "-v", "./tests"}).
		Stdout(ctx)
}

// Returns a container that echoes whatever string argument is provided
func (m *Book) Env(
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		From("golang:latest").
		WithMountedDirectory("/app", source).
		WithWorkdir("/app")
}
