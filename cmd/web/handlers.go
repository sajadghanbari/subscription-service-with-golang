package main

import "net/http"

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	// Handler logic for the home page
	w.Write([]byte("Welcome to the Home Page!"))
}