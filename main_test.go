package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

func TestCallOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"choices": [{"text": "Hello, world!"}]}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := callOpenAI(client, "write a hello world program in Go", server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `{"choices": [{"text": "Hello, world!"}]}
`
	if resp != expected {
		t.Errorf("expected response %q, got %q", expected, resp)
	}
}

func TestCallGemini(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"candidates": [{"content": {"parts": [{"text": "Hello, world!"}]}}]}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := callGemini(client, "write a hello world program in Go", server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `{"candidates": [{"content": {"parts": [{"text": "Hello, world!"}]}}]}
`
	if resp != expected {
		t.Errorf("expected response %q, got %q", expected, resp)
	}
}

func TestCallOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"response": "Hello, world!"}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := callOllama(client, "write a hello world program in Go", server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `{"response": "Hello, world!"}
`
	if resp != expected {
		t.Errorf("expected response %q, got %q", expected, resp)
	}
}
