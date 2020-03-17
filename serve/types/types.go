package types

type TodoCreateInput struct {
	ID       int     `json:"id"`
	ListID   int     `json:"listId"`
	Text     string  `json:"text"`
	Complete bool    `json:"complete"`
	Order    float64 `json:"order"`
}
