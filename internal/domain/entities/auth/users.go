package entities

type User struct {
	ID       int
	Username string
	Email    string
	Password string
	IsActive bool
	RoleId   int
	UserPassword
}

type UserPassword struct {
	PasswordHash string
	PasswordSalt string
}
