package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	// Generate ECDH P-256 key pair
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Get raw bytes
	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := privateKey.PublicKey().Bytes()

	// Encode to base64 URL-safe format
	privateKeyB64 := base64.RawURLEncoding.EncodeToString(privateKeyBytes)
	publicKeyB64 := base64.RawURLEncoding.EncodeToString(publicKeyBytes)

	fmt.Printf("VAPID Keys Generated:\n")
	fmt.Printf("PUSH_VAPID_PUBLIC_KEY=%s\n", publicKeyB64)
	fmt.Printf("PUSH_VAPID_PRIVATE_KEY=%s\n", privateKeyB64)
	fmt.Printf("PUSH_VAPID_SUBJECT=mailto:your-email@example.com\n")
}
