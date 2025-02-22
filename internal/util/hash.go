package util

import (
	hash "crypto/sha256"
	"encoding/base64"
	"io"
)

func HashReader(r io.Reader) (string, error) {
	h := hash.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}
