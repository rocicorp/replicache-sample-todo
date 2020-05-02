package types

type Todo struct {
	ID       int     `json:"id"`
	ListID   int     `json:"listId"`
	Text     string  `json:"text"`
	Complete bool    `json:"complete"`
	Order    float64 `json:"order"`
}

type TodoList struct {
	ID          int `json:"id"`
	OwnerUserID int `json:"ownerUserID"`
}

type TodoCreateInput Todo

type TodoMarkComplete struct {
	ID       int  `json:"id"`
	Complete bool `json:"complete"`
}

type LoginInput struct {
	Email string `json:"email"`
}

type LoginOutput struct {
	Id int `json:"id"`
}
