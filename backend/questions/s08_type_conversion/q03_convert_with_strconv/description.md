Now use the `strconv` library! Create two functions:
1. `IntToStr(n int) string` — converts an integer to a string using `strconv.Itoa`
2. `StrToInt(s string) int` — converts a string to an integer using `strconv.Atoi` (ignore the error)

Expected functions:
func IntToStr(n int) string {}
func StrToInt(s string) int {}

Examples:
IntToStr(42) -> "42"
StrToInt("123") -> 123
