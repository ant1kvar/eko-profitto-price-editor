package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	username  string
	password  string
	sessions  = make(map[string]bool)
	tmplFuncs = template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}
)

func main() {
	_ = godotenv.Load()
	username = os.Getenv("USER_NAME")
	password = os.Getenv("PASSWORD")

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
	tmpl := template.Must(
		template.New("edit.html").Funcs(tmplFuncs).ParseFiles("templates/edit.html"),
	)

	if r.Method == http.MethodPost {
		newTable := r.FormValue("html")

		// Загружаем оригинальный HTML с FTP
		fullHTML, err := LoadHTMLFromFTP()
		if err != nil {
			http.Error(w, "Ошибка загрузки: "+err.Error(), 500)
			return
		}

		// Заменяем только <table id="PriceTable">...</table> в HTML
		start := strings.Index(fullHTML, `<table class="table table-bordered" id="PriceTable"`)
		if start == -1 {
			http.Error(w, "Таблица не найдена", 500)
			return
		}
		end := strings.Index(fullHTML[start:], "</table>")
		if end == -1 {
			http.Error(w, "Ошибка: </table> не найден", 500)
			return
		}
		end += start + len("</table>")
		modifiedHTML := fullHTML[:start] + newTable + fullHTML[end:]

		// Извлекаем и обновляем обе таблицы (desktop + mobile)
		periods, headers, data := ExtractTableData(modifiedHTML)
		finalHTML := UpdateTable(modifiedHTML, periods, headers, data)

		// Сохраняем обратно на FTP
		err = SaveHTMLToFTP(finalHTML)
		if err != nil {
			http.Error(w, "Ошибка сохранения: "+err.Error(), 500)
			return
		}

		// Лог
		user := username
		ip := r.RemoteAddr
		now := time.Now().Format("2006-01-02 15:04:05")
		println("[СОХРАНЕНО] " + now + " — user: '" + user + "' — ip: " + ip)

		http.Redirect(w, r, "/?success=1", http.StatusSeeOther)
		return
	}

	// GET-запрос — отображаем редактор
	fullHTML, err := LoadHTMLFromFTP()
	if err != nil {
		http.Error(w, "Ошибка загрузки: "+err.Error(), 500)
		return
	}
	start := strings.Index(fullHTML, `<table class="table table-bordered" id="PriceTable"`)
	if start == -1 {
		http.Error(w, "Таблица не найдена", 500)
		return
	}
	end := strings.Index(fullHTML[start:], "</table>")
	if end == -1 {
		http.Error(w, "Ошибка: </table> не найден", 500)
		return
	}
	end += start + len("</table>")
	table := fullHTML[start:end]

	data := map[string]interface{}{"HTML": table}
	tmpl.Execute(w, data)
}
