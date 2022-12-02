package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// Check if a file path was provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path as an argument.")
		return
	}

	// Get the file path from the first command-line argument
	filePath := os.Args[1]
	fmt.Printf("Reading from file: %s\n", filePath)

	// Read the entire file into a string
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	// Filter the lines in the file
	filteredLines := filterLines(string(fileContent))

	// Print the filtered lines
	fmt.Println("Filtered lines:")
	for _, line := range filteredLines {
		fmt.Println(line)
	}
}

// Define a type for the tuple containing the line number and the line
type Line struct {
	LineNumber int
	Line       string
}

// filterLines takes an input string with multiple lines and returns all the lines that start with "//gpt"
func filterLines(input string) []Line {
	// Split the input string into separate lines
	lines := strings.Split(input, "\n")

	// Create a slice to hold the filtered lines
	var filteredLines []Line

	// Iterate over the lines
	for i, line := range lines {
		// Trim leading and trailing whitespace from the line
		line = strings.TrimSpace(line)

		// Check if the line starts with "//gpt"
		if strings.HasPrefix(line, "//gpt") {
			// Remove the leading "//gpt" from the line
			line = strings.TrimPrefix(line, "//gpt")

			// Add the line (without the leading "//gpt") to the filtered lines slice
			filteredLines = append(filteredLines, Line{
				LineNumber: i,
				Line:       line,
			})
		}
	}

	// Return the filtered lines
	return filteredLines
}
