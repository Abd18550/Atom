Create a function that shifts each lowercase letter in a string by `n` positions in the alphabet. Non-lowercase characters stay the same. Wrap around from 'z' to 'a'.

Expected function:
func CaesarCipher(s string, shift int) string {}

Examples:
CaesarCipher("abc", 1) -> "bcd"
CaesarCipher("xyz", 3) -> "abc"
CaesarCipher("hello", 0) -> "hello"
