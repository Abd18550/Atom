Create a function that counts the number of words in a sentence using the `strings` package.

Hint: Use `strings.Fields(s)` which splits by whitespace and returns a slice of words.

Expected function:
func CountWords(s string) int {}

Examples:
CountWords("Hello World") -> 2
CountWords("Go is awesome") -> 3
CountWords("") -> 0
