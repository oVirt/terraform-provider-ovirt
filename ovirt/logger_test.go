package ovirt

import (
	"context"
	"fmt"
	"testing"

	log "github.com/ovirt/go-ovirt-client-log/v3"
)

func newTestLogger(t *testing.T) log.Logger {
	return &testLogger{
		t:       t,
		backend: log.NewTestLogger(t),
	}
}

type testLogger struct {
	t       *testing.T
	ctx     context.Context
	backend log.Logger
}

func (t testLogger) Debugf(format string, args ...interface{}) {
	if t.ctx == nil {
		panic(fmt.Errorf("bug: the Terraform logger was used without calling WithContext"))
	}
	t.backend.Debugf(format, args...)
}

func (t testLogger) Infof(format string, args ...interface{}) {
	if t.ctx == nil {
		panic(fmt.Errorf("bug: the Terraform logger was used without calling WithContext"))
	}
	t.backend.Infof(format, args...)
}

func (t testLogger) Warningf(format string, args ...interface{}) {
	if t.ctx == nil {
		panic(fmt.Errorf("bug: the Terraform logger was used without calling WithContext"))
	}
	t.backend.Warningf(format, args...)
}

func (t testLogger) Errorf(format string, args ...interface{}) {
	if t.ctx == nil {
		panic(fmt.Errorf("bug: the Terraform logger was used without calling WithContext"))
	}
	t.backend.Errorf(format, args...)
}

func (t testLogger) WithContext(ctx context.Context) log.Logger {
	return &testLogger{
		t:       t.t,
		ctx:     ctx,
		backend: t.backend,
	}
}
