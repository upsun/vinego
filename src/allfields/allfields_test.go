package allfields

import (
	"testing"

	"github.com/platformsh/vinego/testutils"
)

func TestAnalyzers(t *testing.T) {
	allFieldsAnalyzer := New()
	testutils.RunTests(t, allFieldsAnalyzer)
}
