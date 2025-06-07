package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Get environment variables
	inst := os.Getenv("SPANNER_INSTANCE_ID")
	proj := os.Getenv("SPANNER_PROJECT_ID")
	db := os.Getenv("SPANNER_DATABASE_ID")

	if inst == "" || proj == "" || db == "" {
		log.Fatal("SPANNER_INSTANCE_ID, SPANNER_PROJECT_ID, and SPANNER_DATABASE_ID environment variables must be set")
	}

	// Create the database path
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", proj, inst, db)

	// Connect to the Spanner emulator
	// Use the SPANNER_EMULATOR_HOST environment variable if set, otherwise default to localhost:9010
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		emulatorHost = "localhost:9010"
	}

	client, err := spanner.NewClient(ctx, dbPath,
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
		option.WithEndpoint(emulatorHost),
	)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer client.Close()

	// Execute a simple query to verify the connection
	stmt := spanner.Statement{SQL: "SELECT 1 as test"}
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	// Read the result
	row, err := iter.Next()
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	var test int64
	if err := row.Columns(&test); err != nil {
		log.Fatalf("Failed to parse result: %v", err)
	}

	log.Println("Successfully connected to Spanner emulator and executed query")
}