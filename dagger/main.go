package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

const (
	goVersion           = "1.22.2"
	golangciLintVersion = "v1.57.2"
)

type Ci struct {
	// +private
	Source *Directory
}

func New(
	// Project source directory.
	// +optional
	source *Directory,

	// Checkout the repository (at the designated ref) and use it as the source directory instead of the local one.
	// +optional
	ref string,
) (*Ci, error) {
	if source == nil && ref != "" {
		source = dag.Git("https://github.com/sagikazarmark/demo-kcd-romania-2024.git", GitOpts{
			KeepGitDir: true,
		}).Ref(ref).Tree()
	}

	if source == nil {
		return nil, errors.New("either source or ref is required")
	}

	return &Ci{
		Source: source,
	}, nil
}

func (m *Ci) Build(
	// +optional
	platform Platform,
) *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithWorkdir("/work").
		WithMountedDirectory("/work", m.Source).
		With(func(c *Container) *Container {
			if platform != "" {
				segments := strings.SplitN(string(platform), "/", 3)

				c = c.
					WithEnvVariable("GOOS", segments[0]).
					WithEnvVariable("GOARCH", segments[1])

				if len(segments) > 2 {
					c = c.WithEnvVariable("GOARM", segments[2])
				}
			}

			return c
		}).
		WithExec([]string{"mkdir", "build"}).
		WithExec([]string{"go", "build", "-trimpath", "-o", "build/app", "."})
}

func (m *Ci) Test() *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithWorkdir("/work").
		WithMountedDirectory("/work", m.Source).
		WithExec([]string{"go", "test", "-v", "./..."})
}

func (m *Ci) Lint() *Container {
	return dag.Container()
}

func (m *Ci) Ci(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(func(ctx context.Context) error {
		_, err := m.Build("").Sync(ctx)

		return err
	})

	p.Go(func(ctx context.Context) error {
		_, err := m.Test().Sync(ctx)

		return err
	})

	p.Go(func(ctx context.Context) error {
		_, err := m.Lint().Sync(ctx)

		return err
	})

	return p.Wait()
}
