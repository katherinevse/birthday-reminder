package auth

//go:generate mockgen -source=contract.go -destination=mocks/mockTokenManager.go
type TokenManager interface {
	GenerateJWT(userID int, secretKey string) (string, error)
	ParseJWT(tokenStr string, secretKey string) (*Claims, error)
}
