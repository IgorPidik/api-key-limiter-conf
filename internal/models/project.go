package models

type Project struct {
	ID        string
	Name      string
	UserID    string
	AccessKey string
	Configs   []Config
}
