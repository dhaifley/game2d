package cache

// KeyAccount returns a cache key to be used for account values.
func KeyAccount(id string) string {
	return "Account::" + id
}

// KeyUser returns a cache key to be used for user values.
func KeyUser(id string) string {
	return "User::" + id
}

// KeyAuthToken returns a cache key to be used for authentication token values.
func KeyAuthToken(token string) string {
	return "Token::Auth::" + token
}

// KeyToken returns a cache key to be used for token values.
func KeyToken(token string) string {
	return "Token::" + token
}

// KeyGame returns a cache key to be used for game values.
func KeyGame(id string) string {
	return "Game::" + id
}
