Create a function that takes an age (integer) and validates it. Return an empty string if valid (0-150), or an error message if invalid.

Expected function:
func ValidateAge(age int) string {}

Examples:
ValidateAge(25) -> ""
ValidateAge(-1) -> "age cannot be negative"
ValidateAge(200) -> "age too large"
