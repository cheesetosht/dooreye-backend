package main

import (
	"context"
	"hime-backend/db"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("!! usage: go run ./db/schema.go <path-to-schema.sql>")
	}

	schemaFilePath := os.Args[1]
	sqlBytes, err := os.ReadFile(schemaFilePath)
	if err != nil {
		log.Fatalf("!! unable to read file: \n%v", err)
		os.Exit(1)
	}
	sql := string(sqlBytes)
	db.InitPG()

	_, err = db.PGPool.Exec(context.Background(), sql)

	if err != nil {
		log.Fatalf("!! failed to execute schema file: \n%v", err)
	}

	defer db.ClosePG()
}
