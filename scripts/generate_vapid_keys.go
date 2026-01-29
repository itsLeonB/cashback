package main

import (
	"fmt"
	"log"

	"github.com/itsLeonB/cashback/internal/core/service/webpush"
)

func main() {
	privateKey, publicKey, err := webpush.GenerateKeys()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("VAPID Keys Generated:\n")
	fmt.Printf("PUSH_VAPID_PUBLIC_KEY=%s\n", publicKey)
	fmt.Printf("PUSH_VAPID_PRIVATE_KEY=%s\n", privateKey)
	fmt.Printf("PUSH_VAPID_SUBJECT=mailto:your-email@example.com\n")
}
