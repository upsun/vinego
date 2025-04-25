package allfields

import (
	"testing"

	"github.com/upsun/vinego/src/testutils"
)

func TestAnalyzers(t *testing.T) {
	allFieldsAnalyzer := New()
	testutils.RunTests(t, allFieldsAnalyzer, nil)
}
