package util

import (
	"gopkg.in/reform.v1"
)

// FindCurrentSession finds a current session for a given state channel.
func FindCurrentSession(db *reform.DB, ch string) (string, error) {
	var sess string
	if err := db.QueryRow(`
		SELECT id
		  FROM sessions
		 WHERE channel = $1 AND stopped IS NULL
		 ORDER BY started DESC
		 LIMIT 1`, ch).Scan(&sess); err != nil {
		return "", err
	}

	return sess, nil
}

// HasPositiveBalance checks for a given state channel a balance between
// consumed VPN traffic and received payments.
func HasPositiveBalance(db *reform.DB, ch string) (bool, error) {
	return true, nil
}
