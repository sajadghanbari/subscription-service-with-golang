package main

import "net/http"

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	// Handle login logic here
	// For now, just redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) LogoutPage(w http.ResponseWriter, r *http.Request) {
	// Handle logout logic here
	// For now, just redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "register.page.gohtml", nil)
}
func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	// Handle registration logic here
	// For now, just redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// Handle account activation logic here
	// For now, just redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}