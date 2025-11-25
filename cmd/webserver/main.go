package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"domain_scanner/internal/database"
	"domain_scanner/internal/web"
)

func main() {
	// Command line flags
	port := flag.String("port", ":8080", "Web server port")
	dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.Parse()

	// Validate database URL
	if *dbURL == "" {
		*dbURL = "postgres://scanner:scanner123@localhost:5432/domainscanner?sslmode=disable"
		log.Printf("No DATABASE_URL provided, using default: %s", *dbURL)
	}

	// Connect to database
	db, err := database.New(*dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// Create and start web server
	server := web.NewServer(db)

	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Domain Scanner Web Server v1.4.0              ║")
	fmt.Println("║                                                            ║")
	fmt.Printf("║  Server running at: http://localhost%s                  ║\n", *port)
	fmt.Println("║                                                            ║")
	fmt.Println("║  Features:                                                 ║")
	fmt.Println("║  ✓ Web UI for domain scanning                              ║")
	fmt.Println("║  ✓ Database storage                                        ║")
	fmt.Println("║  ✓ Real-time results                                       ║")
	fmt.Println("║  ✓ Search and filtering                                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if err := server.Start(*port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

