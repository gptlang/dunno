package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
)

const SYSTEM_PROMPT = `
You are a command line assistant. The user will give you a command in natural language and you will return the proper command line syntax. For example, if the user types "create a file named foo.txt", you will return "touch foo.txt".
DO NOT RESPOND WITH ANYTHING OTHER THAN THE COMMAND.

OS: ` + runtime.GOOS + `
Arch: ` + runtime.GOARCH + `
`

type completions struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}
type message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

var HOME_DIR string

func init() {
	var err error
	HOME_DIR, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

}

func main() {
	// Handle ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			os.Exit(0)
		}
	}()
	prompt := get_prompt()

	messages := append(get_history(), message{
		Content: prompt,
		Role:    "user",
	})
	// Construct body
	body := completions{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		Temperature: 0,
	}
	body_json, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	// Construct the request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body_json))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	if os.Getenv("OPENAI_KEY") != "" {
		req.Header.Add("Authorization", "Bearer "+os.Getenv("OPENAI_KEY"))
		set_key(os.Getenv("OPENAI_KEY"))
	} else if get_key() != "" {
		req.Header.Add("Authorization", "Bearer "+get_key())
	} else {
		panic("No API key found")
	}
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// Decode the response to json
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	command := response["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	println(command)

	// Get user input
	var input string
	_, err = fmt.Scanln(&input)
	if err != nil {

	} else if input == "x" {
		os.Exit(0)
	}

	// Execute the command with shell
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		println(err.Error())
	}

	messages = append(messages, message{
		Content: command,
		Role:    "assistant",
	})

	save_history(messages)
}

func get_prompt() string {
	if len(os.Args) == 1 {
		println("Usage: dunno <prompt>")
		println("Example: dunno list tcp network connections")
		println("Enter x to interrupt or enter to allow the assistant to run the command")
		os.Exit(0)
	}
	// Parse arguments and join them with a space
	args := os.Args[1:]
	joined := strings.Join(args, " ")
	return joined
}

func get_key() string {
	mkdir()
	// Read ~/.config/dunno/api_key
	file, err := os.Open(HOME_DIR + "/.config/dunno/api_key")
	if err != nil {
		panic(err)
	}
	// Read the file
	var key string
	err = json.NewDecoder(file).Decode(&key)
	if err != nil {
		panic(err)
	}
	return key
}

func set_key(key string) {
	mkdir()
	// Write to ~/.config/dunno/api_key
	file, err := os.OpenFile(HOME_DIR+"/.config/dunno/api_key", os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	// Encode the key
	err = json.NewEncoder(file).Encode(key)
	if err != nil {
		panic(err)
	}
}

func get_history() []message {
	base_message := []message{
		{
			Content: SYSTEM_PROMPT,
			Role:    "system",
		},
		{
			Content: "List files",
			Role:    "user",
		},
		{
			Content: "ls",
			Role:    "assistant",
		},
		{
			Content: "list tcp network connections",
			Role:    "user",
		},
		{
			Content: "netstat -atn | grep 'tcp'",
			Role:    "assistant",
		},
	}
	if mkdir() {
		return base_message
	}
	// Read history file
	file, err := os.Open(HOME_DIR + "/.config/dunno/history.json")
	if err != nil {
		panic(err)
	}
	// Map to []message
	var history []message
	err = json.NewDecoder(file).Decode(&history)
	if err != nil {
		return base_message
	}
	return history
}
func save_history(history []message) {
	mkdir()
	// Check length of history
	if len(history) > 50 {
		// Truncate history to 50 without removing the first message
		history = history[len(history)-50:]
	}
	// Write to history file
	file, err := os.OpenFile(HOME_DIR+"/.config/dunno/history.json", os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	// Encode the history
	err = json.NewEncoder(file).Encode(history)
	if err != nil {
		panic(err)
	}
}

func mkdir() bool {
	// Check if ~/.config/dunno/history.json exists
	if _, err := os.Stat(HOME_DIR + "/.config/dunno/history.json"); err != nil {
		// Make directory if it doesn't exist
		if _, err := os.Stat(HOME_DIR + "/.config/dunno"); err != nil {
			if _, err := os.Stat(HOME_DIR + "/.config"); err != nil {
				os.Mkdir(HOME_DIR+"/.config", 0755)
			}
			os.Mkdir(HOME_DIR+"/.config/dunno", 0755)
		}
		// If not, create it
		os.Create(HOME_DIR + "/.config/dunno/history.json")
		return true
	}
	if _, err := os.Stat(HOME_DIR + "/.config/dunno/api_key"); err != nil {
		// Create the file
		os.Create(HOME_DIR + "/.config/dunno/api_key")
	}
	return false
}
