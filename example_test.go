package idea_test

import (
	"encoding/hex"
	"fmt"

	idea "github.com/the-hotmann/go-idea"
)

func Example() {
	key := []byte("go-idea-demo-key") // IDEA keys are exactly 16 bytes

	block, err := idea.NewCipher(key)
	if err != nil {
		panic(err)
	}

	plaintext := []byte("IDEA 8b!") // one 8-byte block
	ciphertext := make([]byte, block.BlockSize())
	block.Encrypt(ciphertext, plaintext)
	fmt.Println("encrypted:", hex.EncodeToString(ciphertext))

	decrypted := make([]byte, block.BlockSize())
	block.Decrypt(decrypted, ciphertext)
	fmt.Printf("decrypted: %s\n", decrypted)

	// Output:
	// encrypted: d268905198223171
	// decrypted: IDEA 8b!
}
