package model

type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
	City      string `json:"city"`
}
