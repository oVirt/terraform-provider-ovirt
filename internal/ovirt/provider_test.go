package ovirt

import (
	"testing"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := New()().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
