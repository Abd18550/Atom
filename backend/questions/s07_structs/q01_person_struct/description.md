Define a struct named `Person` with two fields: `Name` (string) and `Age` (int). Then create a function that returns a new `Person` instance.

Expected struct and function:
type Person struct { ... }
func NewPerson(n string, a int) Person {}

Examples:
NewPerson("Bob", 30) -> Person{Name: "Bob", Age: 30}
