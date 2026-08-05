Define an interface `Shape` with a method `Area() int`.

Then create a `Square` struct with a `Side` field (int) that implements the `Shape` interface.

Also create: `func NewSquare(side int) Square`

Expected:
type Shape interface { Area() int }
type Square struct { ... }
func (s Square) Area() int {}
func NewSquare(side int) Square {}

Examples:
s := NewSquare(5)
s.Area() -> 25
