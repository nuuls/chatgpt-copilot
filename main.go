package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

const apiEndpoint = "https://chat.openai.com/backend-api/conversation"

// Payload represents the payload to be sent to the API endpoint
type Payload struct {
	Action          string    `json:"action"`
	Messages        []Message `json:"messages"`
	ParentMessageID string    `json:"parent_message_id"`
	Model           string    `json:"model"`
}

// Message represents a message in the payload
type Message struct {
	ID      string      `json:"id"`
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Content represents the content of a message in the payload
type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

func newPayload(parts string) *Payload {
	// Create a new Payload
	p := &Payload{
		Action: "next",
		Messages: []Message{
			{
				ID:   "7a571da6-a5d8-4724-b7fc-430565618c9e",
				Role: "user",
				Content: &Content{
					ContentType: "text",
					Parts: []string{
						parts,
					},
				},
			},
		},
		ParentMessageID: "a4a85e2c-61ee-4907-9a83-ed1155e784de",
		Model:           "text-davinci-002-render",
	}

	return p
}

func callAPI() {
	// Define the payload
	payload := Payload{
		Action:          "next",
		Messages:        []Message{},
		ParentMessageID: "",
		Model:           "",
	}

	// Add the message to the payload
	// payload.Messages = append(payload.Messages, createMessage())

	// Set the parent message ID and model in the payload
	payload.ParentMessageID = "a4a85e2c-61ee-4907-9a83-ed1155e784de"
	payload.Model = "text-davinci-002-render"

	// Marshal the payload into JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		// Handle error
		return
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		// Handle error
		return
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Handle error
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return
	}

	// Parse the response
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	// Process the response
	// (You can replace this with your own code to process the response)
	fmt.Println("Response:", response)
}

// Response represents the response from the API
type Response struct {
	Message        ResMessage `json:"message"`
	ConversationID string     `json:"conversation_id"`
	Error          error      `json:"error"`
}

// Message represents a message in the response
type ResMessage struct {
	ID         string                 `json:"id"`
	Role       string                 `json:"role"`
	User       interface{}            `json:"user"`
	CreateTime interface{}            `json:"create_time"`
	UpdateTime interface{}            `json:"update_time"`
	Content    ResContent             `json:"content"`
	EndTurn    interface{}            `json:"end_turn"`
	Weight     float64                `json:"weight"`
	Metadata   map[string]interface{} `json:"metadata"`
	Recipient  string                 `json:"recipient"`
}

// Content represents the content of a message in the response
type ResContent struct {
	ContentType string        `json:"content_type"`
	Parts       []interface{} `json:"parts"`
}

// Define a function for decoding an EventSource http.Response
func decodeEventSourceResponse(resp *http.Response) (*Response, error) {
	// Check the response status code to make sure it's 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	// Check the Content-Type header to make sure it's an EventSource response
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		return nil, fmt.Errorf("unexpected Content-Type: %s", contentType)
	}

	// Read the response body as a byte slice
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Split the response body into separate lines
	lines := bytes.Split(responseBody, []byte("\n"))

	// Loop through the lines and decode the last non-empty line
	var response *Response
	for i := len(lines) - 1; i >= 0; i-- {
		// Skip empty lines
		if len(lines[i]) == 0 {
			continue
		}

		// Check if the line is "data: [DONE]"
		if bytes.Equal(lines[i], []byte("data: [DONE]")) {
			// Skip the "data: [DONE]" line
			continue
		}

		// Strip the "data: " prefix from the line
		line := bytes.TrimPrefix(lines[i], []byte("data: "))

		// Decode the line into a Response object
		response = &Response{}
		if err := json.Unmarshal(line, response); err != nil {
			return nil, err
		}

		// Stop the loop when a non-empty line is successfully decoded
		break
	}

	return response, nil
}
