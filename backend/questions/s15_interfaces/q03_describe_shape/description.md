Using the Shape interface from the previous exercise, create a function that takes a Shape and returns a description string.

Expected function:
func DescribeShape(s Shape) string {}

The function should return: "This shape has area: X" where X is the area.

Examples:
s := NewSquare(4)
DescribeShape(s) -> "This shape has area: 16"
