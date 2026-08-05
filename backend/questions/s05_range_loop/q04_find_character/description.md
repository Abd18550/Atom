Create a function that takes a string and a target character (`rune` type). Use a `range` loop to find and return the index of the first occurrence. If not found, return -1.

Expected function:
func FindCharIndex(s string, target rune) int {}

Examples:
FindCharIndex("hello", 'e') -> 1
FindCharIndex("world", 'z') -> -1
