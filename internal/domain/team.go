package domain

type Team struct {
	TeamName string
	Members  []TeamMember
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}
