Define a struct `Rectangle` with `Width` and `Height` as integers. Add a method to the struct that calculates and returns its area.

Expected:
type Rectangle struct { ... }
func (r Rectangle) Area() int {}

Examples:
r := Rectangle{Width: 5, Height: 2}
r.Area() -> 10
