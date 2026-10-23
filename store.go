// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	contextsDirName = "_mycontext"
	markerFilename  = ".mycontext-branch"
	localDirName    = ".local"
)

// store manages the contexts of one Git repository,
// stored in <repo>/../_mycontext/<repo dir name>/<branch>.
type store struct {
	repoDir string
	root    string
}

func newStore(dir string) (*store, error) {
	top, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, errors.New("not inside a Git repository")
	}
	top = filepath.FromSlash(top)
	return &store{
		repoDir: top,
		root:    filepath.Join(filepath.Dir(top), contextsDirName, filepath.Base(top)),
	}, nil
}

func (s *store) currentBranch() (string, error) {
	b, err := git(s.repoDir, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		return "", errors.New("HEAD is detached; check out a branch first")
	}
	return b, nil
}

func (s *store) branches() (map[string]bool, error) {
	out, err := git(s.repoDir, "for-each-ref", "--format=%(refname)", "refs/heads")
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool)
	for ref := range strings.FieldsSeq(out) {
		m[strings.TrimPrefix(ref, "refs/heads/")] = true
	}
	// An unborn branch (no commits yet) is not listed by for-each-ref.
	if cur, err := s.currentBranch(); err == nil {
		m[cur] = true
	}
	return m, nil
}

func (s *store) dir(branch string) (string, error) {
	p := filepath.FromSlash(branch)
	if !filepath.IsLocal(p) || filepath.Clean(p) == "." {
		return "", fmt.Errorf("invalid branch name %q", branch)
	}
	return filepath.Join(s.root, p), nil
}

func (s *store) exists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, markerFilename))
	return err == nil
}

func (s *store) path(local bool) (string, error) {
	branch, err := s.currentBranch()
	if err != nil {
		return "", err
	}
	dir, err := s.dir(branch)
	if err != nil {
		return "", err
	}
	if !s.exists(dir) {
		if err := s.create(dir); err != nil {
			return "", err
		}
	}
	if local {
		dir = filepath.Join(dir, localDirName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func (s *store) create(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	gitignore := filepath.Join(filepath.Dir(s.root), ".gitignore")
	if _, err := os.Stat(gitignore); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(gitignore, []byte(localDirName+"/\n"), 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(dir, markerFilename), nil, 0o644)
}

// contexts returns the branch names of all contexts, sorted.
func (s *store) contexts() ([]string, error) {
	var branches []string
	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == s.root && errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			return err
		}
		if d.IsDir() && d.Name() == localDirName {
			return fs.SkipDir
		}
		if !d.IsDir() && d.Name() == markerFilename {
			rel, err := filepath.Rel(s.root, filepath.Dir(path))
			if err != nil {
				return err
			}
			branches = append(branches, filepath.ToSlash(rel))
		}
		return nil
	})
	return branches, err
}

func (s *store) list(out io.Writer) error {
	contexts, err := s.contexts()
	if err != nil {
		return err
	}
	branches, err := s.branches()
	if err != nil {
		return err
	}
	cur, _ := s.currentBranch()
	for _, c := range contexts {
		prefix, suffix := "  ", ""
		if c == cur {
			prefix = "* "
		}
		if !branches[c] {
			suffix = " (gone)"
		}
		fmt.Fprintf(out, "%s%s%s\n", prefix, c, suffix)
	}
	return nil
}

func (s *store) remove(branch string) error {
	dir, err := s.dir(branch)
	if err != nil {
		return err
	}
	if !s.exists(dir) {
		return fmt.Errorf("no context for branch %q", branch)
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	s.removeEmptyParents(dir)
	return nil
}

func (s *store) move(from, to string) error {
	fromDir, err := s.dir(from)
	if err != nil {
		return err
	}
	toDir, err := s.dir(to)
	if err != nil {
		return err
	}
	if !s.exists(fromDir) {
		return fmt.Errorf("no context for branch %q", from)
	}
	if _, err := os.Stat(toDir); err == nil {
		return fmt.Errorf("context for branch %q already exists", to)
	}
	if err := os.MkdirAll(filepath.Dir(toDir), 0o755); err != nil {
		return err
	}
	if err := os.Rename(fromDir, toDir); err != nil {
		return err
	}
	s.removeEmptyParents(fromDir)
	return nil
}

func (s *store) prune(dryRun bool, out io.Writer) error {
	contexts, err := s.contexts()
	if err != nil {
		return err
	}
	branches, err := s.branches()
	if err != nil {
		return err
	}
	for _, c := range contexts {
		if branches[c] {
			continue
		}
		if dryRun {
			fmt.Fprintf(out, "would remove %s\n", c)
			continue
		}
		dir, err := s.dir(c)
		if err != nil {
			return err
		}
		// May already be gone with a removed parent context.
		if !s.exists(dir) {
			continue
		}
		if err := s.remove(c); err != nil {
			return err
		}
		fmt.Fprintf(out, "removed %s\n", c)
	}
	return nil
}

// removeEmptyParents removes empty directories from dir's parent up to and including s.root.
func (s *store) removeEmptyParents(dir string) {
	stop := filepath.Dir(s.root)
	for d := filepath.Dir(dir); d != stop; d = filepath.Dir(d) {
		if os.Remove(d) != nil {
			return
		}
	}
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
