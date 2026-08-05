package database

import "backend/models"

func getStages() []models.Stage {
	return []models.Stage{
		{Title: "Variables & Types", Description: "Learn Go data types: string, int, bool, float64, rune, and constants.", OrderIndex: 1},
		{Title: "Control Flow", Description: "Master if/else conditions and comparison logic.", OrderIndex: 2},
		{Title: "For Loop", Description: "Classic counted loop iterations.", OrderIndex: 3},
		{Title: "While Loop", Description: "Condition-based loops using for.", OrderIndex: 4},
		{Title: "Range Loop", Description: "Iterate over slices and strings with range.", OrderIndex: 5},
		{Title: "Slices & Arrays", Description: "Working with dynamic lists of items.", OrderIndex: 6},
		{Title: "Structs", Description: "Custom composite data types and methods.", OrderIndex: 7},
		{Title: "Type Conversion", Description: "Converting between types manually and with strconv.", OrderIndex: 8},
		{Title: "Functions Advanced", Description: "Multiple returns, variadic functions, and more.", OrderIndex: 9},
		{Title: "Strings & Libraries", Description: "The strings package and text manipulation.", OrderIndex: 10},
		{Title: "Maps", Description: "Key-value data structures for lookups.", OrderIndex: 11},
		{Title: "OS Args & Input", Description: "Reading command-line arguments with os.Args.", OrderIndex: 12},
		{Title: "Error Handling", Description: "Handling errors the Go way.", OrderIndex: 13},
		{Title: "Pointers", Description: "Understanding memory addresses and references.", OrderIndex: 14},
		{Title: "Interfaces", Description: "Defining behavior contracts for types.", OrderIndex: 15},
		{Title: "Combined Challenges", Description: "Problems that combine multiple concepts.", OrderIndex: 16},
		{Title: "Mini Project: CLI Tools", Description: "Build command-line tools using os.Args and strconv.", OrderIndex: 17},
		{Title: "Mini Project: Text Processor", Description: "Process and analyze text data.", OrderIndex: 18},
	}
}
