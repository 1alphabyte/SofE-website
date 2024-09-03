package main

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

type changePwdData struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type GetData struct {
	Key   string `clover:"_id" json:"key"`
	Ip    string `clover:"IP" json:"ip"`
	Name  string `clover:"name" json:"name"`
	Email string `clover:"email" json:"email"`
	Msg   string `clover:"msg" json:"msg"`
	Lname string `clover:"lname" json:"lname"`
	Time  int64  `clover:"time" json:"time"`
}
