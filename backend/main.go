package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
)

type dbData struct {
	Key   string   `json:"key"`
	Value FormData `json:"value"`
}

type FormData struct {
	Name  string `json:"name"`
	Lname string `json:"lname" omit_empty:"true"`
	Email string `json:"email"`
	Msg   string `json:"msg"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.URL.Query().Has("etkn") {
		http.Error(w, "Password not provided", http.StatusUnauthorized)
		return
	}
	if r.URL.Query().Get("etkn") != "Supersecurevalue" {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}
	var data FormData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
	}

	var re = regexp.MustCompile(`^[^@]+@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$/i`)
	if !re.MatchString(data.Email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if data.Name == "" || data.Email == "" || data.Msg == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	d, err := deta.New()
	if err != nil {
		http.Error(w, "failed to create deta client", http.StatusInternalServerError)
	}
	db, err := base.New(d, "testing")
	if err != nil {
		http.Error(w, "failed to create deta base", http.StatusInternalServerError)
	}

	tmp := &dbData{
		Key:   data.Name,
		Value: data,
	}
	_, err = db.Put(tmp)
	if err != nil {
		http.Error(w, "failed to put data", http.StatusInternalServerError)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler)

	log.Printf("App listening on port %s!", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
