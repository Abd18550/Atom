Create a function that takes a string. If it can be converted to a positive integer, return that integer and an empty string. Otherwise return 0 and an appropriate error message.

Expected function:
func ParsePositive(s string) (int, string) {}

Examples:
ParsePositive("42") -> 42, ""
ParsePositive("-5") -> 0, "not positive"
ParsePositive("abc") -> 0, "invalid number"
