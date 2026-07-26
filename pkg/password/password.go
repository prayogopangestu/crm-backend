// Package password wraps the bcrypt password hashing primitives so callers
// do not depend on golang.org/x/crypto/bcrypt directly. Centralising the
// hashing here keeps cost configuration and error mapping in one place.
package password

import "golang.org/x/crypto/bcrypt"

// Hash returns the bcrypt hash of `plain` using the provided cost. A cost of
// bcrypt.DefaultCost is a good default for most deployments.
func Hash(plain string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare reports whether `plain` matches the previously hashed `hashed`.
// It returns a non-nil error when the password is wrong or the hash is invalid.
func Compare(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
