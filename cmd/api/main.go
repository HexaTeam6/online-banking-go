package main

import (
	"fmt"
	"os"

	// Import all dependencies to ensure they are tracked in go.mod
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/leanovate/gopter"
	_ "github.com/rs/zerolog"
	_ "github.com/stretchr/testify/assert"
	_ "go.uber.org/mock/mockgen/model"
	_ "golang.org/x/crypto/bcrypt"
	_ "golang.org/x/time/rate"
)

func main() {
	fmt.Println("Online Banking API Server")
	os.Exit(0)
}
