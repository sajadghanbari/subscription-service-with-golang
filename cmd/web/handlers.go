package main

import (
	"fmt"
	"html/template"
	"net/http"
	"subscription-service/data"
)

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil{
		app.ErrorLog.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user , err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(),"error","Invalid credentials")
		http.Redirect(w,r,"/login",http.StatusSeeOther)
		return 
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		app.Session.Put(r.Context(),"error","Invalid credentials")
		http.Redirect(w,r,"/login",http.StatusSeeOther)
		return 
	}

	if !validPassword{
		msg := Message{
			To:email,
			Subject: "failed login in attempt",
			Data: "Invalid login attempt!",
		}

		app.sendEmail(msg)

		app.Session.Put(r.Context(),"error","Invalid credentials")
		http.Redirect(w,r,"/login",http.StatusSeeOther)
		return 
	}

	app.Session.Put(r.Context(),"userID",user.ID)
	app.Session.Put(r.Context(),"user",user)

	app.Session.Put(r.Context(),"flash","successful login")


	http.Redirect(w,r,"/",http.StatusSeeOther)

}

func (app *Config) LogoutPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	http.Redirect(w,r,"/login",http.StatusSeeOther)


}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w,r, "register.page.gohtml", nil)
}
func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil{
		app.ErrorLog.Println(err)
	}
	//validate data

	//create a user
	u := data.User{
		Email: r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName: r.Form.Get("last-name"),
		Password: r.Form.Get("password"),
		Active: 0,
		IsAdmin: 0,
	}

	_, err = u.Insert(u)
	if err != nil {
		app.Session.Put(r.Context(),"error","Unable to create user.")
		http.Redirect(w,r,"/register",http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("http://localhost:8080/activate?email=%s",u.Email)
	signedURL := GenerateTokenFromString(url)
	app.InfoLog.Println(signedURL)

	msg := Message{
		To: u.Email,
		Subject: "Activate account",
		Template: "confirm-email",
		Data: template.HTML(signedURL),
	}

	app.sendEmail(msg)
	app.Session.Put(r.Context(),"flash","Confirmation Email sent")
	http.Redirect(w,r,"/login",http.StatusSeeOther)

	
	
}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	// Handle account activation logic here
	// For now, just redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}