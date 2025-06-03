package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	username string
	password string
	sessions = make(map[string]bool)
)

func main() {
	_ = godotenv.Load()
	username = os.Getenv("USER")
	password = os.Getenv("PASSWORD")
	fmt.Println("USER:", username)
	fmt.Println("PASSWORD:", password)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/", authMiddleware(editHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	if r.Method == http.MethodPost {
		userInput := r.FormValue("username")
		passInput := r.FormValue("password")
		ip := r.RemoteAddr
		now := time.Now().Format("2006-01-02 15:04:05")

		if userInput == username && passInput == password {
			sessionID := uuid.New().String()
			sessions[sessionID] = true
			http.SetCookie(w, &http.Cookie{
				Name:   "session",
				Value:  sessionID,
				Path:   "/",
				MaxAge: 1800,
			})
			logLogin(now, userInput, ip, true)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		logLogin(now, userInput, ip, false)
		tmpl.Execute(w, "Неверный логин или пароль")
		return
	}

	tmpl.Execute(w, nil)
}

func logLogin(timestamp, user, ip string, success bool) {
	status := "НЕУДАЧНО"
	if success {
		status = "УСПЕШНО"
	}
	log := "[" + status + "] " + timestamp + " — user: '" + user + "' — ip: " + ip
	println(log)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session")
	if err == nil {
		delete(sessions, c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil || !sessions[c.Value] {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/edit.html"))
	data := map[string]string{
		"HTML": `<table><tr><td>Здесь будет таблица с FTP</td></tr></table>`,
	}
	tmpl.Execute(w, data)
}
