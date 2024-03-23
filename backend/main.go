package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/google/uuid"
)

type dbData struct {
	Key   string   `json:"key,omitempty"`
	Value FormData `json:"value"`
}

type FormData struct {
	Name  string `json:"name"`
	Lname string `json:"lname,omitempty"`
	Email string `json:"email"`
	Msg   string `json:"msg"`
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Response struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Expires int64  `json:"__expires,omitempty"`
}

func setUpDetaBase(name string) (*base.Base, error) {
	d, err := deta.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create deta client")
	}
	db, err := base.New(d, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create deta base")
	}
	return db, nil
}

func signIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "Password not provided", http.StatusUnauthorized)
		return
	}
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
	}
	db, err := setUpDetaBase("auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var storedCreds Response
	err = db.Get(creds.Username, &storedCreds)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid username", http.StatusForbidden)
		return
	}
	if storedCreds.Value != creds.Password {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(72 * time.Hour).Unix()
	db.Put(&Response{
		Key:     sessionToken,
		Value:   creds.Username,
		Expires: expiresAt,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionToken": sessionToken,
		"expiresAt":    expiresAt,
	})
}

func checkCookie(r *http.Request) int {
	tkn, err := r.Cookie("token")
	if err != nil || tkn == nil || tkn.Value == "" {
		return http.StatusUnauthorized
	}
	db, err := setUpDetaBase("auth")
	if err != nil {
		return http.StatusInternalServerError
	}
	err = db.Get(tkn.Value, nil)
	if err != nil {
		return http.StatusForbidden
	}
	return http.StatusOK
}

func add(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.URL.Query().Has("etkn") {
		http.Error(w, "Password not provided", http.StatusUnauthorized)
		return
	}
	if r.URL.Query().Get("etkn") != "95d9d334b7dc7fd211b3" {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}
	var data FormData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "failed to parse body", http.StatusBadRequest)
	}

	var re = regexp.MustCompile(`^[^@]+@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$`)
	if !re.MatchString(data.Email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if data.Name == "" || data.Email == "" || data.Msg == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	db, err := setUpDetaBase("contact-form")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Put(&dbData{
		Value: data,
	})
	if err != nil {
		http.Error(w, "failed to put data", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Success!"))
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := checkCookie(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	db, err := setUpDetaBase("contact-form")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []dbData
	_, err = db.Fetch(&base.FetchInput{
		Dest: &results,
	})
	if err != nil {
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(results)

}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := checkCookie(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := setUpDetaBase("contact-form")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	key := string(bodyBytes)
	err = db.Delete(key)
	if err != nil {
		http.Error(w, "failed to delete data", http.StatusInternalServerError)
		return
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/add", add)
	http.HandleFunc("/get", get)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/signin", signIn)

	log.Printf("App listening on port %s!", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
