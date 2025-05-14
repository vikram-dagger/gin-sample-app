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
		WithMountedDirectory("/app", source.WithoutDirectory(".dagger")).
		WithWorkdir("/app")

	diff := ctr.
		WithExec([]string{"sh", "-c", "git diff > /tmp/a.diff"}).
		File("/tmp/a.diff")

	env := dag.Env(dagger.EnvOpts{Privileged: true}).
		WithDirectoryInput("source", source, "directory with source code").
		WithFileInput("diff", diff, "file with code diff").
		WithFileOutput("after", "updated changelog file")

	prompt := `
		- You are an expert in the Go programming language.
		- You are also an expert in the Gin framework and database integrations.
		- You have access to a directory with source code and an OpenAPI spec.
		- The directory has tools to let you read and write files.
		- You also have access to a diff file with code changes.
		- Understand the changes by reading the source code, the diff and the OpenAPI spec.
		- Update the changelog file in the source directory with a summary of the changes and return the updated file.
		- Focus only on the Go files in the directory.
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
