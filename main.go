package main

import (
	"auth-session/connection"
	"context"
	"fmt"

	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnection()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public/"))))

	route.HandleFunc("/", homePage).Methods("GET")
	route.HandleFunc("/contact", contactPage).Methods("GET")
	route.HandleFunc("/project", projectPage).Methods("GET")
	route.HandleFunc("/project", addProject).Methods("POST")
	route.HandleFunc("/project/{id}", detailProject).Methods("GET")
	route.HandleFunc("/editProject/{id}", editProject).Methods("GET")
	route.HandleFunc("/updateProject/{id}", updateProject).Methods("POST")
	route.HandleFunc("/deleteProject/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/register", registerPage).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/login", loginPage).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server running on port:5050")
	http.ListenAndServe("localhost:5050", route)
}

var Data = map[string]interface{}{
	"IsLogin": true,
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	rows, errQuery := connection.Conn.Query(context.Background(), "SELECT * FROM tb_project")
	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	var result []dataReceive

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	for rows.Next() {
		var each = dataReceive{}

		err := rows.Scan(&each.ID, &each.Projectname, &each.Startdate, &each.Enddate, &each.Description, &each.Technologies)
		if err != nil {
			fmt.Println("Message : " + err.Error())
			return
		}

		each.Duration = countduration(each.Startdate, each.Enddate)

		result = append(result, each)
	}

	dataCaller := map[string]interface{}{
		"Projects": result,
		"Data":     Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataCaller)
}

func projectPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/myProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	dataCaller := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataCaller)
}

// Struct menentukan variable dan tipe datanya
type dataReceive struct {
	ID           int
	Projectname  string
	Description  string
	Technologies []string
	Startdate    time.Time
	Enddate      time.Time
	Duration     string
}

func countduration(start time.Time, end time.Time) string {
	distance := end.Sub(start)

	//Ubah milisecond menjadi bulan, minggu dan hari
	monthDistance := int(distance.Hours() / 24 / 30)
	weekDistance := int(distance.Hours() / 24 / 7)
	daysDistance := int(distance.Hours() / 24)

	// variable buat menampung durasi yang sudah diolah
	var duration string
	// pengkondisian yang akan mengirimkan durasi yang sudah diolah
	if monthDistance >= 1 {
		duration = strconv.Itoa(monthDistance) + " months"
	} else if monthDistance < 1 && weekDistance >= 1 {
		duration = strconv.Itoa(weekDistance) + " weeks"
	} else if monthDistance < 1 && daysDistance >= 0 {
		duration = strconv.Itoa(daysDistance) + " days"
	} else {
		duration = "0 days"
	}

	return duration
}

func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10485760) // menggunakan r.ParseMultipartForm karena mengizinkan pengiriman file beda dengan r.ParseForm yang tidak mengizinkan pengiriman gambar
	// 10485760 byte = 10mb

	if err != nil {
		log.Fatal(err)
	}

	projectname := r.PostForm.Get("project-name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"] // pakai r.Form karena ingin menangkap query string

	const timeFormat = "2006-01-02"
	startDate, _ := time.Parse(timeFormat, r.PostForm.Get("start-date"))
	endDate, _ := time.Parse(timeFormat, r.PostForm.Get("end-date"))

	_, insertRow := connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(project_name, start_date, end_date, description, technologies) VALUES ($1, $2, $3, $4, $5)", projectname, startDate, endDate, description, technologies)
	if insertRow != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/project-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	var resultData = dataReceive{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", ID).Scan(&resultData.ID, &resultData.Projectname, &resultData.Startdate, &resultData.Enddate, &resultData.Description, &resultData.Technologies)

	resultData.Duration = countduration(resultData.Startdate, resultData.Enddate)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	detailProject := map[string]interface{}{
		"Projects": resultData,
		"Data":     Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, detailProject)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, deleteRows := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if deleteRows != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + deleteRows.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/editProject.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	var editData = dataReceive{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", ID).Scan(&editData.ID, &editData.Projectname, &editData.Startdate, &editData.Enddate, &editData.Description, &editData.Technologies)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	dataEdit := map[string]interface{}{
		"Projects": editData,
		"Data":     Data,
	}

	tmpl.Execute(w, dataEdit)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10485760)

	if err != nil {
		log.Fatal(err)
	}

	projectname := r.PostForm.Get("project-name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"] // pakai r.Form karena ingin menangkap query string

	const timeFormat = "2006-01-02"
	startDate, _ := time.Parse(timeFormat, r.PostForm.Get("start-date"))
	endDate, _ := time.Parse(timeFormat, r.PostForm.Get("end-date"))
	ID, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, updateRow := connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5 WHERE id = $6", projectname, startDate, endDate, description, technologies, ID)
	if updateRow != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	dataCaller := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataCaller)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	dataCaller := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataCaller)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")

	password := r.PostForm.Get("password")
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, insertRow := connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if insertRow != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	dataCaller := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dataCaller)
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Options.MaxAge = 10800

	session.AddFlash("Login success", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Options.MaxAge = -1

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
