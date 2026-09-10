// adminhash reads a password from stdin so it never appears in process arguments.
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"os"
	"strings"
)

func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 100))
	if err != nil {
		panic(err)
	}
	password := strings.TrimSpace(string(raw))
	if len(password) < 16 || len(password) > 72 {
		fmt.Fprintln(os.Stderr, "Password must be 16–72 bytes")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(hash))
}
