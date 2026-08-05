Create a function that returns the number of unique characters in a string.

Hint: Use a `map[rune]bool` to track seen characters.

Expected function:
func UniqueCount(s string) int {}

Examples:
UniqueCount("hello") -> 4  // h, e, l, o
UniqueCount("aaa") -> 1
UniqueCount("abcabc") -> 3
