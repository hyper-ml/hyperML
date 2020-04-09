package qs

import "github.com/hyper-ml/hyperml/server/pkg/db"

// IsErrRecNotFound : Custom error returned Database when key doesnt exist
func IsErrRecNotFound(e error) bool {
	return db.IsErrRecNotFound(e)
}
