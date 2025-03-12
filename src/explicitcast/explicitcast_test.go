package explicitcast

import (
	"testing"

	"github.com/upsun/vinego/src/testutils"
)

func TestAnalyzers(t *testing.T) {
	testutils.RunTests(t, New())
}
