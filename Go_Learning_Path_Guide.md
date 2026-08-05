# The Go Journey: Complete Learning Path

## Stage 1: Variables & Types
*Basics of Go data types and variables.*

### 1. Hello Variables
Create a string variable named `greeting` and assign it the value "Hello, Go!". Print the variable to the console.

**Expected function:**
```go
func main() {}
```
**Examples:**
```text
// Output:
Hello, Go!
```

### 2. Sum Function
Create a function that takes two integers and returns their sum.

**Expected function:**
```go
func Add(a, b int) int {}
```
**Examples:**
```text
Add(5, 5) -> 10
Add(-2, 3) -> 1
```

### 3. Constant Value
Create a constant named `Pi` and assign it the value 3.14. Print the constant to the console.

**Expected function:**
```go
func main() {}
```
**Examples:**
```text
// Output:
3.14
```

---

## Stage 2: Control Flow
*Conditions and comparison logic.*

### 1. Adult Check
Create a function that takes an age as an integer and returns true if the person is 18 or older, and false otherwise.

**Expected function:**
```go
func IsAdult(age int) bool {}
```
**Examples:**
```text
IsAdult(20) -> true
IsAdult(15) -> false
IsAdult(18) -> true
```

### 2. Sign Check
Create a function that takes an integer and returns "Positive" if it is greater than zero, "Negative" if it is less than zero, and "Zero" if it is exactly zero.

**Expected function:**
```go
func CheckSign(n int) string {}
```
**Examples:**
```text
CheckSign(5) -> "Positive"
CheckSign(-3) -> "Negative"
CheckSign(0) -> "Zero"
```

### 3. FizzBuzz
Create a function that takes an integer. If the number is divisible by 3 and 5, return "FizzBuzz". If divisible by only 3, return "Fizz". If divisible by only 5, return "Buzz". Otherwise, return an empty string "".

**Expected function:**
```go
func FizzBuzz(n int) string {}
```
**Examples:**
```text
FizzBuzz(15) -> "FizzBuzz"
FizzBuzz(3) -> "Fizz"
FizzBuzz(5) -> "Buzz"
FizzBuzz(2) -> ""
```

---

## Stage 3: For Loop
*Standard for loop iterations.*

### 1. Print 1 to 5
Use a for loop to print the numbers from 1 to 5, each on a new line.

**Expected function:**
```go
func main() {}
```
**Examples:**
```text
// Output:
1
2
3
4
5
```

### 2. Sum of N
Create a function that takes an integer `n` and returns the sum of all numbers from 1 up to `n`.

**Expected function:**
```go
func Sum(n int) int {}
```
**Examples:**
```text
Sum(5) -> 15   // (1+2+3+4+5)
Sum(3) -> 6    // (1+2+3)
```

### 3. Factorial Calculation
Create a function that takes a positive integer `n` and returns its factorial (the product of all positive integers less than or equal to `n`). Return 1 if `n` is 0.

**Expected function:**
```go
func Factorial(n int) int {}
```
**Examples:**
```text
Factorial(5) -> 120
Factorial(4) -> 24
```

---

## Stage 4: While Loop
*Simulating while loops in Go.*

### 1. Halve Until Zero
Create a function that takes an integer `n`. Using a while-style loop (`for n > 0`), divide the number by 2 in each iteration and count the number of divisions it takes to reach 0. Return the count.

**Expected function:**
```go
func CountDivisions(n int) int {}
```
**Examples:**
```text
CountDivisions(10) -> 4  // 10->5->2->1->0
CountDivisions(1) -> 1
```

### 2. Sum Until Limit
Create a function that takes an integer `limit`. Using a while-style loop, keep adding numbers (1 + 2 + 3 + ...) until the sum is strictly greater than or equal to the `limit`. Return the last number added.

**Expected function:**
```go
func SumUntil(limit int) int {}
```
**Examples:**
```text
SumUntil(10) -> 4  // 1+2+3+4=10
SumUntil(20) -> 6  // +5=15, +6=21
```

### 3. Digit Counter
Create a function that takes a positive integer `n` and returns the number of digits in it using a while-style loop.

**Expected function:**
```go
func CountDigits(n int) int {}
```
**Examples:**
```text
CountDigits(123) -> 3
CountDigits(4567) -> 4
```

---

## Stage 5: Range Loop
*Iterating safely over collections.*

### 1. Print Characters
Define a string `word := "Go!"` and use a `range` loop to print each character on a new line. (Hint: Use `string(c)` to convert the rune into a readable character).

**Expected function:**
```go
func main() {}
```
**Examples:**
```text
// Output:
G
o
!
```

### 2. Sum ASCII Values
Create a function that takes a string and uses a `range` loop to calculate and return the sum of the byte/ASCII values of all its characters.

**Expected function:**
```go
func SumASCII(s string) int {}
```
**Examples:**
```text
SumASCII("ABC") -> 198  // (65 + 66 + 67)
SumASCII("a") -> 97
```

### 3. Find Character
Create a function that takes a string and a target character (`rune` type). Use a `range` loop to find and return the index (position) of the first occurrence of that character. If not found, return -1.

**Expected function:**
```go
func FindCharIndex(s string, target rune) int {}
```
**Examples:**
```text
FindCharIndex("hello", 'e') -> 1
FindCharIndex("world", 'z') -> -1
```

---

## Stage 6: Slices & Arrays
*Lists of items.*

### 1. Append Element
Create a function that takes a slice of integers and a new integer, appends the new integer to the slice, and returns the modified slice.

**Expected function:**
```go
func AppendElement(arr []int, n int) []int {}
```
**Examples:**
```text
AppendElement([]int{1, 2}, 3) -> []int{1, 2, 3}
```

### 2. Contains
Create a function that takes an array of integers and a target number, and checks if the array contains this number.

**Expected function:**
```go
func Contains(a []int, t int) bool {}
```
**Examples:**
```text
Contains([]int{1, 2}, 2) -> true
Contains([]int{1, 2}, 3) -> false
```

### 3. Reverse
Create a function that takes a slice of integers and returns a new slice with the elements in reverse order.

**Expected function:**
```go
func Reverse(a []int) []int {}
```
**Examples:**
```text
Reverse([]int{1, 2, 3}) -> []int{3, 2, 1}
Reverse([]int{5, 10}) -> []int{10, 5}
```

---

## Stage 7: Structs
*Custom composite types.*

### 1. Person Struct
Define a struct named `Person` with two fields: `Name` (string) and `Age` (int). Then create a function that returns a new `Person` instance using the provided arguments.

**Expected struct and function:**
```go
type Person struct { ... }
func NewPerson(n string, a int) Person {}
```
**Examples:**
```text
NewPerson("Bob", 30) -> Person{Name: "Bob", Age: 30}
```

### 2. Point Dist
Define a struct `Point` with `X` and `Y` as integers. Create a function that calculates the squared distance between two points: (p1.X - p2.X)^2 + (p1.Y - p2.Y)^2

**Expected struct and function:**
```go
type Point struct { ... }
func DistanceSquared(p1, p2 Point) int {}
```
**Examples:**
```text
DistanceSquared(Point{0,0}, Point{3,4}) -> 25
```

### 3. Rect Area
Define a struct `Rectangle` with `Width` and `Height` as integers. Add a method to the struct that calculates and returns its area.

**Expected struct and method:**
```go
type Rectangle struct { ... }
func (r Rectangle) Area() int {}
```
**Examples:**
```text
r := Rectangle{Width: 5, Height: 2}
r.Area() -> 10
```

---

## Stage 8: Final Challenges
*Putting concepts together.*

### 1. Palindrome
Create a function that checks if a given string is a palindrome (reads the same forwards and backwards). Assume the string contains only lowercase ASCII characters.

**Expected function:**
```go
func IsPalindrome(s string) bool {}
```
**Examples:**
```text
IsPalindrome("racecar") -> true
IsPalindrome("hello") -> false
IsPalindrome("madam") -> true
```

### 2. Count Vowels
Create a function that counts and returns the number of vowels ('a', 'e', 'i', 'o', 'u') in a given string.

**Expected function:**
```go
func CountVowels(s string) int {}
```
**Examples:**
```text
CountVowels("hello") -> 2
CountVowels("apple") -> 2
CountVowels("xyz") -> 0
```

### 3. Min Max
Create a function that takes a slice of integers and returns two values: the minimum and the maximum elements in the slice. Assume the slice is never empty.

**Expected function:**
```go
func FindMinMax(a []int) (min int, max int) {}
```
**Examples:**
```text
FindMinMax([]int{3, 1, 5, 2}) -> 1, 5
FindMinMax([]int{10, 10}) -> 10, 10
```
