package db

import (
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"

	"ewallet-backend/utils"

	"github.com/lestrrat-go/jwx/jwk"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitPostgres() {
	var err error

	dbUser := utils.GetEnv("DB_USER", "ewallet_user")
	dbPass := utils.GetEnv("DB_PASS", "ewallet_pass")
	dbHost := utils.GetEnv("DB_HOST", "localhost")
	dbPort := utils.GetEnv("DB_PORT", "5432")
	dbName := utils.GetEnv("DB_NAME", "ewallet_db")

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database (Ping failed): %v", err)
	}

	fmt.Println("✅ Connected to PostgreSQL")
}

func GetUserPublicKey(userID string) (*rsa.PublicKey, error) {
	var pubKeyJWKString string
	err := DB.QueryRow(`SELECT "public_key" FROM "users" WHERE "id" = $1`, userID).Scan(&pubKeyJWKString)
	if err != nil {
		log.Printf("error di getuserpublic key: %s for userID: '%s'", err.Error(), userID)
		return nil, err
	}

	key, err := jwk.ParseKey([]byte(pubKeyJWKString))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWK string for user %s: %w", userID, err)
	}

	var raw any
	if err := key.Raw(&raw); err != nil {
		return nil, fmt.Errorf("failed to get raw key from JWK for user %s: %w", userID, err)
	}

	pubKey, ok := raw.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("converted key for user %s is not an RSA public key, got %T", userID, raw)
	}
	return pubKey, nil
}
