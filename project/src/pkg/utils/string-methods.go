package utils

import (
	"math/rand"
	"time"
)

// RandomString returns a random string given a desired length
func RandomString(length uint16) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// create a random source
	source := rand.NewSource(time.Now().UnixNano())
	// create a byte slice of the given length
	b := make([]byte, length)

	// fill the slice with random characters from the charset
	for i := range b {
		b[i] = charset[source.Int63()%int64(len(charset))]
	}
	// return the string representation of the slice
	return string(b)
}
