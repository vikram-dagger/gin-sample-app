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

func (m *Book) Changelog(
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.File {
	ctr := dag.Container().
		From("golang:latest").
		WithMountedDirectory("/app", source).
		WithWorkdir("/app")

	env := dag.Env(dagger.EnvOpts{Privileged: true}).
		WithContainerInput("before", ctr, "the container with the source code").
		WithFileOutput("after", "the changelog file with the updated changelog")

	prompt := `
		- You are an expert in the Go programming language.
		- You are also an expert in the Gin framework and database integrations.
		- You have access to a container with the code in the /app directory.
		- The container has tools to let you read and write the code and obtain a diff.
		- Obtain a diff and analyze the changes in the code.
		- Compare the changes with the OpenAPI spec in the /app/openapi.yml file.
		- In the container, update the changelog with a summary of the changes.
		- Be sure to always write your changes to the container.
		- Focus only on Go files within the /app directory.
	`

	work := dag.LLM().
		WithEnv(env).
		WithPrompt(prompt)

	return work.Env().Output("after").AsFile()
}

func (m *Book) Env(
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		From("golang:latest").
		WithMountedDirectory("/app", source).
		WithWorkdir("/app")
}
