Convert a single digit (0-9) to its string character WITHOUT using any library like strconv or fmt.

Hint: In Go, '0' is a rune with value 48. Adding an integer n to '0' gives the rune for that digit. Use string() to convert.

Expected function:
func DigitToChar(n int) string {}

Examples:
DigitToChar(0) -> "0"
DigitToChar(5) -> "5"
DigitToChar(9) -> "9"
