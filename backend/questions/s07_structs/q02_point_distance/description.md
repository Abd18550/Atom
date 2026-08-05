Define a struct `Point` with `X` and `Y` as integers. Create a function that calculates the squared distance between two points: (p1.X - p2.X)^2 + (p1.Y - p2.Y)^2

Expected:
type Point struct { ... }
func DistanceSquared(p1, p2 Point) int {}

Examples:
DistanceSquared(Point{0,0}, Point{3,4}) -> 25
