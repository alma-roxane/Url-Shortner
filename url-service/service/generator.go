package service

import (
	"crypto/rand"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Generator struct {
	length int
}

func NewGenerator(length int) *Generator {
	if length < 5 {
		length = 7
	}
	return &Generator{length: length}
}

func (g *Generator) Generate() (string, error) {
	buf := make([]byte, g.length)
	raw := make([]byte, g.length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range raw {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}
