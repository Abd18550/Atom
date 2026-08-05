Define a struct `Car` with a `Brand` field (string). Add a method `Greet()` that returns "Vroom! I'm a " followed by the brand name.

Also create a constructor function `NewCar(brand string) Car`.

Expected:
type Car struct { ... }
func (c Car) Greet() string {}
func NewCar(brand string) Car {}

Examples:
c := NewCar("Toyota")
c.Greet() -> "Vroom! I'm a Toyota"
