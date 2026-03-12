package adapters

import (
	"crypto/md5"
	"encoding/hex"
)

type HasherRepositoryMD5 struct{}

func NewHasherRepositoryMD5() *HasherRepositoryMD5 {
	return &HasherRepositoryMD5{}
}

func (h *HasherRepositoryMD5) Hash(raw string) string {
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (h *HasherRepositoryMD5) Compare(raw, encoded string) bool {
	return h.Hash(raw) == encoded
}
