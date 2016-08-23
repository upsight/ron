package color

import "fmt"

// Red wraps the input string in ansii red color escape codes.
func Red(input string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", input)
}

// Green wraps the input string in ansii green color escape codes.
func Green(input string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", input)
}

// Yellow wraps the input string in ansii yellow color escape codes.
func Yellow(input string) string {
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", input)
}

// Blue wraps the input string in ansii blue color escape codes.
func Blue(input string) string {
	return fmt.Sprintf("\x1b[34m%s\x1b[0m", input)
}
