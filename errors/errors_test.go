package errors_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dhaifley/game2d/errors"
)

func TestNew(t *testing.T) {
	t.Parallel()

	a := errors.New(errors.ErrServer, "test", "test", "test")
	if a.Code.Name != "Server" {
		t.Errorf("Expected code: Server, got: %v", a.Code.Name)
	}

	host, _ := os.Hostname()
	if a.Svr != host {
		t.Errorf("Expected server: %v, got: %v", host, a.Svr)
	}

	if a.Proc != "testing.tRunner" {
		t.Errorf("Expected procedure: testing.tRunner, got: %v", a.Proc)
	}

	if len(a.Data) != 1 {
		t.Errorf("Expected data length: 1, got: %v", len(a.Data))
	}

	if a.Data["test"] != "test" {
		t.Errorf("Expected test data: test, got: %v", a.Data["test"])
	}
}

func TestContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		a := errors.Context(ctx)

		if a.Code.Name != "Canceled" {
			t.Errorf("Expected code: Canceled, got: %v", a.Code.Name)
		}
	default:
		t.Fatal("Context did not properly cancel")
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	a := errors.New(errors.ErrServer, "test")

	b := errors.Wrap(a, errors.ErrForbidden, "test2", "test", "test")

	if b.Code.Name != "Server" {
		t.Errorf("Expected code: Server, got: %v", b.Code.Name)
	}

	if b.Err.Error() != a.Error() {
		t.Errorf("Expected error: %v, got: %v",
			a.Error(), b.Err.Error())
	}

	host, _ := os.Hostname()
	if a.Svr != host {
		t.Errorf("Expected server: %v, got: %v", host, a.Svr)
	}

	if b.Time != a.Time {
		t.Errorf("Expected time: %v, got: %v", a.Time, b.Time)
	}

	if len(b.Data) != 1 {
		t.Errorf("Expected data length: 1, got: %v", len(b.Data))
	}

	if b.Data["test"] != "test" {
		t.Errorf("Expected test data: test, got: %v", b.Data["test"])
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	e := errors.New(errors.ErrForbidden, "unauthorized", "test", "test")

	e.Time = 0

	exp := `{"code":"Forbidden","status":403,`

	if !strings.Contains(e.String(), exp) {
		t.Errorf("Expected string to contain: %v, got: %v",
			exp, e.String())
	}

	exp = `"data":{"test":"test"}`

	if !strings.Contains(e.String(), exp) {
		t.Errorf("Expected string to contain: %v, got: %v",
			exp, e.String())
	}
}
