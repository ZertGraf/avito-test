package domain

type User struct {
	UserID   string
	Username string
	TeamName string // связь с командой
	IsActive bool   // можно ли назначить ревьювером
}
