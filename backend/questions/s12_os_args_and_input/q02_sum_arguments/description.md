Write a program that takes any number of numeric arguments and prints their sum.

Hint: Loop through `os.Args[1:]` and use `strconv.Atoi` to convert each to int.

Expected:
func main() {}

Examples:
$ go run main.go 1 2 3
6

$ go run main.go 10 20
30
