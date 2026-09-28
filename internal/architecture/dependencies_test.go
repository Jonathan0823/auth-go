package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestCoreDoesNotImportOuterLayers(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test source")
	}
	coreDir := filepath.Join(filepath.Dir(source), "..", "core")
	forbidden := []string{
		"github.com/Jonathan0823/auth-go/internal/adapter",
		"github.com/Jonathan0823/auth-go/internal/bootstrap",
		"github.com/Jonathan0823/auth-go/internal/config",
		"github.com/Jonathan0823/auth-go/internal/observability",
		"github.com/Jonathan0823/auth-go/internal/platform",
	}

	err := filepath.WalkDir(coreDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			for _, prefix := range forbidden {
				if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
					t.Errorf("core source %s imports outer-layer package %q", path, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
