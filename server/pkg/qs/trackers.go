package qs

import "fmt"

// getPodsTracker : Used by pod insert/update triggers.
// Generic flag that allows listening to changes of all pods
func getPodsTracker() string {
	return "Pod{}"
}

// getUserTracker : Used by pod insert/updates triggers
func getUserTracker(username string) string {
	return "User{" + username + "}"
}

// getPodIDTracker : Used by pod insert/updates triggers
func getPodIDTracker(id uint64) string {
	return "Pod{" + fmt.Sprintf("%d", id) + "}"
}
