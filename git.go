package registry

import (
	"fmt"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

// GitConfig contains the configuration of the Git provider.
type GitConfig struct {
	// URL is the repository URL.
	URL string

	// Branch is the Git branch to use.
	Branch string
}

// NewGitConfig creates a default Git provider configuration pointing to the production branch.
func NewGitConfig() GitConfig {
	return GitConfig{
		URL:    "https://github.com/oasisprotocol/metadata-registry",
		Branch: "production",
	}
}

// NewTestGitConfig creates a Git provider configuration pointing to the test branch.
func NewTestGitConfig() GitConfig {
	return GitConfig{
		URL:    "https://github.com/oasisprotocol/metadata-registry",
		Branch: "testing",
	}
}

// NewGitProvider creates a new git-backed metadata registry provider.
func NewGitProvider(cfg GitConfig) (Provider, error) {
	fs := memfs.New()
	_, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:           cfg.URL,
		Depth:         1,
		ReferenceName: plumbing.NewBranchReferenceName(cfg.Branch),
		SingleBranch:  true,
		Tags:          git.NoTags,
	})
	if err != nil {
		return nil, fmt.Errorf("registry/git: failed to clone repository: %w", err)
	}
	return NewFilesystemProvider(fs)
}
