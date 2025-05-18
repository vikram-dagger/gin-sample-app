package main

import (
	"context"
	"dagger/book/internal/dagger"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v61/github"
	"golang.org/x/oauth2"
)

type Foo struct {
	File *dagger.File
	Data string
}

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

func (m *Book) UpdateChangelog(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	repository string,
	// +optional
	ref string,
	// +optional
	token dagger.Secret,
) *Foo {
	source = source
	ctr := dag.Container().
		From("golang:latest").
		WithMountedDirectory("/app", source).
		WithWorkdir("/app")

	diff := ctr.
		//WithExec([]string{"git", "fetch", "origin", "main"}).
		WithExec([]string{"sh", "-c", "git diff origin/main > /tmp/a.diff"}).
		File("/tmp/a.diff")

	env := dag.Env(dagger.EnvOpts{Privileged: true}).
		WithDirectoryInput("source", source, "directory with source code").
		WithFileInput("diff", diff, "file with code diff").
		WithFileOutput("after", "updated CHANGELOG.md file")

	prompt := `
		- You are an expert in the Go programming language.
		- You are also an expert in the Gin framework and database integrations.
		- You have access to a directory with source code and an OpenAPI spec.
		- The directory has tools to let you read and write files.
		- You also have access to a diff file with code changes.
		- Understand the changes by reading the source code, the diff and the OpenAPI spec.
		- Ignore all changes in the .dagger directory.
		- Update the CHANGELOG.md file in the source directory.
		- When updating the CHANGELOG.md file, increment the version and add a summary of the changes.
		- You must save the CHANGELOG.md file after updating it.
		- You are not done until your changes are saved back to the CHANGELOG.md file.
		- Focus only on the Go files in the directory.
	`

	work := dag.LLM().
		WithEnv(env).
		WithPrompt(prompt)

	changelogFile := work.Env().Output("after").AsFile()

	// Check if we should open a PR
	if repository != "" && ref != "" {
		diffFile := *ctr.
			WithFile("/app/CHANGELOG.md", changelogFile).
			WithExec([]string{"sh", "-c", "git diff CHANGELOG.md > /tmp/changelog.diff"}).
			File("/tmp/changelog.diff")

		prURL, err := OpenPR(ctx, repository, ref, diffFile, token)
		if err != nil {
			panic(fmt.Errorf("failed to open PR: %w", err))
		}
		fmt.Println("PR URL: ", prURL)

		commentURL, err := m.WritePRComment(ctx, repository, ref, fmt.Sprintf("Changelog updated: see PR %s", prURL), token)
		if err != nil {
			panic(fmt.Errorf("failed to write PR comment: %w", err))
		}
		fmt.Println("Comment URL: ", commentURL)

		return &Foo{Data: prURL}
	}

	return &Foo{File: changelogFile}
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

func OpenPR(
	ctx context.Context,
	repository string,
	ref string,
	diffFile dagger.File,
	token dagger.Secret,
) (string, error) {
	// Extract PR number from ref
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(ref)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid ref format: %s", ref)
	}
	prNumber := matches[1]
	newBranch := fmt.Sprintf("patch-from-pr-%s", prNumber)

	// Setup GitHub client
	plaintext, err := token.Plaintext(ctx)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: plaintext})
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	// Split repo
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: %s", repository)
	}
	owner, repo := parts[0], parts[1]

	// Get original PR
	prNumInt := github.Int(mustParseInt(prNumber))
	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, *prNumInt)
	if err != nil {
		return "", fmt.Errorf("failed to get original PR: %w", err)
	}
	baseBranch := pr.GetHead().GetRef()

	// Run container to apply patch
	remoteURL := fmt.Sprintf("https://%s@github.com/%s.git", plaintext, repository)
	diff, err := diffFile.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get file contents: %w", err)
	}
	fmt.Println("Diff: ", diff)
	_, err = dag.Container().
		From("alpine/git").
		WithNewFile("/tmp/changelog.diff", diff).
		WithWorkdir("/app").
		WithExec([]string{"cat", "/tmp/changelog.diff"}).
		/*
			//WithEnvVariable("GITHUB_TOKEN", plaintext).
			WithExec([]string{"git", "init"}).
			WithExec([]string{"git", "branch", "-m", "main"}).
			WithExec([]string{"git", "config", "user.name", "Dagger Agent"}).
			WithExec([]string{"git", "config", "user.email", "vikram@dagger.io"}).
			WithExec([]string{"sh", "-c", "git remote add origin " + remoteURL}).
			WithExec([]string{"git", "fetch", "origin", fmt.Sprintf("pull/%s/head:%s", prNumber, newBranch)}).
			WithExec([]string{"git", "checkout", newBranch}).
			WithExec([]string{"git", "apply", "--allow-empty", "/tmp/changelog.diff"}).
			WithExec([]string{"git", "add", "."}).
			WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Follows up on PR #%s", prNumber)}).
			WithExec([]string{"git", "push", "--set-upstream", "origin", newBranch}).
		*/
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to apply and push changes: %w", err)
	}

	// Create new PR
	newPR := &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Automated follow-up for PR #%s", prNumber)),
		Head:  github.String(fmt.Sprintf("%s:%s", owner, newBranch)),
		Base:  github.String(baseBranch),
		Body:  github.String(fmt.Sprintf("This PR follows up PR #%s using `%s`.", prNumber, newBranch)),
	}
	createdPR, _, err := gh.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return "", fmt.Errorf("failed to create new PR: %w", err)
	}

	return createdPR.GetHTMLURL(), nil
}

func (m *Book) WritePRComment(
	ctx context.Context,
	repository string,
	ref string,
	body string,
	token dagger.Secret,
) (string, error) {
	// Extract PR number using regex
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(ref)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract PR number from ref: %s", ref)
	}
	prNumberStr := matches[1]
	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		return "", fmt.Errorf("invalid PR number: %s", prNumberStr)
	}

	// Extract owner and repo
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: %s", repository)
	}
	owner, repo := parts[0], parts[1]

	// Setup GitHub client
	plaintext, err := token.Plaintext(ctx)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: plaintext})
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	// Create the comment
	comment := &github.IssueComment{
		Body: github.String(body),
	}
	createdComment, _, err := gh.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	if err != nil {
		return "", fmt.Errorf("failed to create comment: %w", err)
	}

	return createdComment.GetHTMLURL(), nil
}

func mustParseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("invalid int: %s", s))
	}
	return v
}
