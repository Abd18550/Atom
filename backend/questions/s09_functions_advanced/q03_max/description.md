Create a function that takes a slice of integers and returns the largest one.

Hint: Use a variadic function `...int` to accept any number of arguments.

Expected function:
func Max(nums ...int) int {}

Examples:
Max(1, 5, 3) -> 5
Max(-1, -5, -2) -> -1
Max(7, 7, 7) -> 7
Max(42) -> 42
Max(10, 20, 30, 40, 50) -> 50
