// Package secureinput provides OS-level masked input for passwords and tokens.
package secureinput

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadLine reads a visible line from stdin after printing prompt.
func ReadLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// ReadMasked reads input from stdin without echoing characters (for passwords/tokens).
// Falls back to visible input when stdin is not a terminal (e.g. test pipes).
func ReadMasked(prompt string) (string, error) {
	fmt.Print(prompt)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	// Fallback: stdin is not a terminal (piped / test mode)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
