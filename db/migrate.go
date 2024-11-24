package db

import (
	"database/sql"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

func AddMigrations(db *sql.DB) {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "1",
				Up: []string{
					`CREATE TYPE role AS ENUM ('creator', 'student', 'admin')`,
				},
				Down: []string{
					`DROP TYPE role`,
				},
			},
			{
				Id: "2",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS "user" (id SERIAL PRIMARY KEY, username VARCHAR(255) UNIQUE, role role DEFAULT('student'), institute VARCHAR(255), verified BOOLEAN DEFAULT(false), faculty VARCHAR(255), department VARCHAR(255), email VARCHAR(255) UNIQUE, password VARCHAR(100))`,
				},
				Down: []string{
					`DROP IF EXISTS "user"`,
				},
			},
		},
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("Couldn't apply the migrations: %s", err)
	}

	log.Printf("Applied %d migrations!", n)
}
