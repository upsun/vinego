package capturederr

import (
	"testing"

	"github.com/platformsh/vinego/testutils"
)

func TestAnalyzers(t *testing.T) {
	testutils.RunTests(t, New())
}
