package request

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/dhaifley/game2d/errors"
)

// Query messages represent query string search requests.
type Query struct {
	Search string `json:"search,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Skip   int64  `json:"skip,omitempty"`
	Sort   string `json:"sort,omitempty"`
}

func NewQuery() *Query {
	return &Query{
		Search: "",
		Size:   100,
		Skip:   0,
		Sort:   "",
	}
}

// ParseQuery parses a string in query string format into a Query value that
// can be used for search functions.
func ParseQuery(values url.Values) (*Query, error) {
	req := &Query{}

	for qk, qv := range values {
		qk = strings.ToLower(qk)

		if len(qv) == 0 {
			continue
		}

		switch qk {
		case "search":
			req.Search = qv[0]
		case "skip":
			if strings.TrimSpace(qv[0]) != "" {
				i, err := strconv.ParseInt(strings.TrimSpace(qv[0]), 10, 64)
				if err != nil || i < 0 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"invalid query skip value",
						"query", values)
				}

				req.Skip = i
			}
		case "size":
			if strings.TrimSpace(qv[0]) != "" {
				i, err := strconv.ParseInt(strings.TrimSpace(qv[0]), 10, 64)
				if err != nil || i < 0 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"invalid query size value",
						"query", values)
				}

				req.Size = i
			}
		case "sort":
			req.Sort = strings.Join(qv, ",")
		}
	}

	return req, nil
}
