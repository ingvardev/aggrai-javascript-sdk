package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ingvar/aiaggregator/packages/adapters"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer pool.Close()

	repo := adapters.NewPostgresTenantOwnerRepository(pool)
	owner, err := repo.GetByEmail(context.Background(), "admin@localhost")
	if err != nil {
		fmt.Println("GetByEmail error:", err)
		return
	}

	fmt.Println("Owner found:")
	fmt.Println("  ID:", owner.ID)
	fmt.Println("  Email:", owner.Email)
	fmt.Println("  PasswordHash:", owner.PasswordHash)
	fmt.Println("  Active:", owner.Active)

	// Test password
	password := "admin123"
	err = bcrypt.CompareHashAndPassword([]byte(owner.PasswordHash), []byte(password))
	if err != nil {
		fmt.Println("Password check FAILED:", err)
	} else {
		fmt.Println("Password check PASSED!")
	}

	// Test CheckPassword method
	if owner.CheckPassword(password) {
		fmt.Println("CheckPassword() PASSED!")
	} else {
		fmt.Println("CheckPassword() FAILED!")
	}
}
