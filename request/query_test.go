package request_test

import (
	"net/url"
	"testing"

	"github.com/dhaifley/game2d/request"
)

func TestParseQuery(t *testing.T) {
	t.Parallel()

	q := "search=test%20(test:test)&skip=10&size=10&sort=test" +
		"&ver=v2&search=(test1:test1)&sort=-test1&summary=test,test1"

	values, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}

	req, err := request.ParseQuery(values)
	if err != nil {
		t.Fatal(err)
	}

	expS := "test (test:test)"

	if req.Search != expS {
		t.Errorf("Expected search: %v, got: %v", expS, req.Search)
	}

	expI := int64(10)

	if req.Size != expI {
		t.Errorf("Expected size: %v, got: %v", expI, req.Size)
	}

	expI = int64(10)

	if req.Skip != expI {
		t.Errorf("Expected skip: %v, got: %v", expI, req.Skip)
	}

	expS = "test,-test1"

	if req.Sort != expS {
		t.Errorf("Expected sort: %v, got: %v", expS, req.Sort)
	}
}
