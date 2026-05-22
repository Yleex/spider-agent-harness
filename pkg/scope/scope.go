package scope

import (
	"fmt"
	"os"
	"path/filepath"
)

type Validator struct {
	ProjectRoot string
}

func New(projectRoot string) (*Validator, error) {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolviendo project root: %w", err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		real = abs
	}
	info, err := os.Stat(real)
	if err != nil {
		return nil, fmt.Errorf("project root no accesible: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project root no es un directorio: %s", real)
	}
	return &Validator{ProjectRoot: real}, nil
}

func (v *Validator) Safe(target string) (string, error) {
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolviendo path: %w", err)
	}

	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		real = abs
	}

	if !hasPrefix(real, v.ProjectRoot) {
		return "", fmt.Errorf("%w: %s", ErrOutsideScope, target)
	}

	return real, nil
}

func (v *Validator) Root() string {
	return v.ProjectRoot
}

func (v *Validator) IsInsideProject(paths ...string) bool {
	for _, p := range paths {
		if _, err := v.Safe(p); err != nil {
			return false
		}
	}
	return true
}

func hasPrefix(p, prefix string) bool {
	p = filepath.Clean(p)
	prefix = filepath.Clean(prefix)
	if len(p) < len(prefix) {
		return false
	}
	if p[:len(prefix)] != prefix {
		return false
	}
	if len(p) > len(prefix) && p[len(prefix)] != os.PathSeparator {
		return false
	}
	return true
}
