package session

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandJwtSecret() string {
	const length = 32
	// Calculate the number of bytes needed (each hex character is 4
	// bits, so 2 hex chars per byte)
	buf_size := length / 2
	if length % 2 != 0 {
		buf_size++ // If the length is odd, add an extra byte to have a full byte
	}
	randomBytes := make([]byte, buf_size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Could not access rand!!") // Should never happend, according to docs
	}

	hexString := hex.EncodeToString(randomBytes)
    //If the length was odd, trim the result
    if length % 2 != 0{
        hexString = hexString[:length]
    }

	return hexString
}

