package genkey

import (
	"crypto/rand"
	"fmt"
	"os"
)

func dieIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] %v\n", err)
		os.Exit(1)
	}
}

func dieWith(fs string, args ...interface{}) {
	outStr := fmt.Sprintf("[!] %s\n", fs)
	fmt.Fprintf(os.Stderr, outStr, args...)
	os.Exit(1)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func MakeKey(keyfile string) {
	key := make([]byte, 64)

	_, err := rand.Read(key)
	if err != nil {
		dieWith("Unable to Generate a random Key!!!")
	}

	if fileExists(keyfile) {
		dieWith("Key already exists! All set...")
	}

	file, err := os.OpenFile(
		keyfile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0440,
	)
	dieIf(err)
	defer file.Close()

	// Write bytes to file
	byteSlice := []byte(key)
	bytesWritten, err := file.Write(byteSlice)
	dieIf(err)

	fmt.Printf("Wrote %d bytes.\n", bytesWritten)

	fmt.Println("Successfully generated key file...")
}
