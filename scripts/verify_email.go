package main

import (
	"context"
	"fmt"
	"os"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run scripts/verify_email.go <email>")
		os.Exit(1)
	}
	email := os.Args[1]

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("load config:", err)
		os.Exit(1)
	}

	pool, err := datastore.NewPool(ctx, cfg)
	if err != nil {
		fmt.Println("create pool:", err)
		os.Exit(1)
	}
	defer pool.Close()

	tag, err := pool.Exec(ctx,
		`UPDATE users SET email_verified = true, email_verified_at = now() WHERE email = $1`,
		email,
	)
	if err != nil {
		fmt.Println("update user:", err)
		os.Exit(1)
	}

	if tag.RowsAffected() == 0 {
		fmt.Println("user not found:", email)
		os.Exit(1)
	}

	_, _ = pool.Exec(ctx,
		`UPDATE email_verifications SET used_at = now() WHERE user_id = (SELECT id FROM users WHERE email = $1)`,
		email,
	)

	fmt.Println("verified:", email)
}
