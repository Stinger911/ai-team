package main

import (
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <model> <task>")
		return
	}

	model := os.Args[1]
	task := os.Args[2]

	fmt.Printf("Selected model: %s\n", model)
	fmt.Printf("Task: %s\n", task)

	client := &http.Client{}

	var response string
	var err error

	switch model {
	case "openai":
		response, err = callOpenAI(client, task, "https://api.openai.com/v1/completions")
	case "gemini":
		response, err = callGemini(client, task, "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent")
	case "ollama":
		response, err = callOllama(client, task, "http://localhost:11434/api/generate")
	default:
		fmt.Println("Invalid model selected. Please choose from: openai, gemini, ollama")
		return
	}

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Response:", response)
}

func callOpenAI(client *http.Client, task string, apiURL string) (string, error) {
	fmt.Println("Calling OpenAI API...")
	apiKey := os.Getenv("OPENAI_API_KEY")

	requestBody := strings.NewReader(`{
		"model": "text-davinci-003",
		"prompt": "` + task + `",
		"max_tokens": 100
	}`)

	req, err := http.NewRequest("POST", apiURL, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}

func callGemini(client *http.Client, task string, apiURL string) (string, error) {
	fmt.Println("Calling Gemini API...")
	apiKey := os.Getenv("GEMINI_API_KEY")

	requestBody := strings.NewReader(`{
		"contents": [{
			"parts":[
				{"text": "` + task + `"}
			]
		}]
	}`)

	req, err := http.NewRequest("POST", apiURL, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = "key=" + apiKey

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}

func callOllama(client *http.Client, task string, apiURL string) (string, error) {
	fmt.Println("Calling Ollama API...")

	requestBody := strings.NewReader(`{
		"model": "llama2",
		"prompt": "` + task + `"
	}`)

	req, err := http.NewRequest("POST", apiURL, requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}
