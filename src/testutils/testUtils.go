package testutils

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func RunTests(t *testing.T, analyzer *analysis.Analyzer) {
	root := "testdata"
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("failed to load test case %s: %s", path, err)
			return nil
		}
		relPath, _ := filepath.Rel(root, path)
		dir, cleanup, err := analysistest.WriteFiles(map[string]string{
			relPath: string(contents),
		})
		if err != nil {
			if err != nil {
				t.Errorf("failed to prep temp test dir: %s", err)
				return nil
			}
		}
		defer cleanup()
		fmt.Printf("========= at %s\n", path)
		results := analysistest.Run(t, dir, analyzer, filepath.Base(filepath.Dir(path)))
		for _, res := range results {
			if res.Err != nil {
				t.Errorf("analyzer failed on %s: %s", path, res.Err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
