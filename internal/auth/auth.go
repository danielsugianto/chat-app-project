// internal/auth/auth.go
package auth

import "errors"

var validUsers = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

func Authenticate(username, password string) error {
	if validUsers[username] != password {
		return errors.New("invalid username or password")
	}
	return nil
}
