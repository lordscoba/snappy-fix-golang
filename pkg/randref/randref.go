package randref

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

func New(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.ToUpper(prefix) + "_" + time.Now().UTC().Format("20060102T150405") + "_" + hex.EncodeToString(b)
}
