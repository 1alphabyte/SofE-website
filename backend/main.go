package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	c "github.com/ostafen/clover"
)

var db, _ = c.Open("data")

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
		log.Println(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
	}

	query := db.Query("auth").Where(c.Field("username").Eq(creds.Username))
	exists, err := query.Exists()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Invalid username", http.StatusForbidden)
		return
	}
	storedCreds, err := query.FindFirst()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if storedCreds.Get("password") != creds.Password {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(72 * time.Hour).Unix()
	s := c.NewDocument()
	s.Set("sessionToken", sessionToken)
	s.Set("expiresAt", expiresAt)
	s.Set("username", creds.Username)
	_, err = db.InsertOne("auth", s)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

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

	data := db.Query("auth").Where(c.Field("sessionToken").Eq(tkn.Value))
	d, err := data.Exists()
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}
	if !d {
		return http.StatusForbidden
	}
	decoded, err := data.FindFirst()
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError
	}

	if decoded.Get("expiresAt").(int64) < time.Now().Unix() {
		data.Delete()
		return http.StatusUnauthorized
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
		log.Println(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
	}

	if data.Name == "" || data.Email == "" || data.Msg == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	var re = regexp.MustCompile(`^[^@]+@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$`)
	if !re.MatchString(data.Email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	doc := c.NewDocument()
	doc.Set("name", data.Name)
	doc.Set("email", data.Email)
	doc.Set("msg", data.Msg)
	doc.Set("lname", data.Lname)
	doc.Set("IP", r.Header.Get("STE-Real-IP"))
	doc.Set("time", time.Now().Unix())

	_, err = db.InsertOne("data", doc)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
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

	docs, err := db.Query("data").FindAll()
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		return
	}
	var results []GetData
	var result GetData
	for _, doc := range docs {
		doc.Unmarshal(&result)
		results = append(results, result)
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	key := string(bodyBytes)
	err = db.Query("data").DeleteById(key)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to delete data", http.StatusInternalServerError)
		return
	}
}

func changePwd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
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
	// parse body
	var b changePwdData
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to parse body", http.StatusBadRequest)
		return
	}
	// check if old and new password are provided
	if b.NewPassword == "" || b.OldPassword == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}
	// check if new password is long enough
	if len(b.NewPassword) <= 8 {
		http.Error(w, "Must be 8 or more characters", http.StatusBadRequest)
		return
	}
	query := db.Query("auth").Where(c.Field("username").Eq("Admin"))
	doc, err := query.FindFirst()
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		return
	}
	if doc.Get("password") != b.OldPassword {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}
	// invalidate all sessions
	err = db.Query("auth").Delete()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	// update password
	d := c.NewDocument()
	d.Set("username", "Admin")
	d.Set("password", b.NewPassword)
	_, err = db.InsertOne("auth", d)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to set password", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success!"))
}

func initDB() {
	db.CreateCollection("data")
	db.CreateCollection("auth")
	doc := c.NewDocument()
	doc.Set("username", "Admin")
	doc.Set("password", "defaultpassword")
	db.InsertOne("auth", doc)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initDB()
		log.Println("Database initialized")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/add", add)
	http.HandleFunc("/api/get", get)
	http.HandleFunc("/api/delete", delete)
	http.HandleFunc("/api/signin", signIn)
	http.HandleFunc("/api/changepwd", changePwd)

	log.Printf("App listening on port %s!", port)
	if err := http.ListenAndServe("127.0.0.1:"+port, nil); err != nil {
		log.Fatal(err)
	}
}
