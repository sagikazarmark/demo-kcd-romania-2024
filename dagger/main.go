package main

import (
	"context"
	"errors"

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
) *File {
	return dag.Go(GoOpts{Version: goVersion}).
		Build(m.Source, GoBuildOpts{
			Platform: string(platform),
			Trimpath: true,
		})
}

func (m *Ci) Test() *Container {
	return dag.Go(GoOpts{Version: goVersion}).
		WithSource(m.Source).
		Exec([]string{"go", "test", "-v", "./..."})
}

func (m *Ci) Lint() *Container {
	return dag.GolangciLint(GolangciLintOpts{
		Version:   golangciLintVersion,
		GoVersion: goVersion,
	}).
		Run(m.Source, GolangciLintRunOpts{
			Verbose: true,
		})
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
