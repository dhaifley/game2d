package cache_test

import (
	"testing"

	"github.com/dhaifley/game2d/cache"
)

func TestCacheKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		exp string
		run func() string
	}{
		{
			exp: "Account::test",
			run: func() string { return cache.KeyAccount("test") },
		},
		{
			exp: "User::test",
			run: func() string { return cache.KeyUser("test") },
		},
		{
			exp: "Token::Auth::test",
			run: func() string { return cache.KeyAuthToken("test") },
		},
		{
			exp: "Token::test",
			run: func() string { return cache.KeyToken("test") },
		},
		{
			exp: "Game::test",
			run: func() string { return cache.KeyGame("test") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.exp, func(t *testing.T) {
			t.Parallel()

			res := tt.run()

			if tt.exp != res {
				t.Errorf("Expected key: %v, got: %v", tt.exp, res)
			}
		})
	}
}
