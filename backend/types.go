package main

type dbData struct {
	Key   string   `json:"key,omitempty"`
	Value FormData `json:"value"`
	IP    string   `json:"ip"`
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

type changePwdData struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}
