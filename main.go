package main


import (
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"github.com/DylanCoon99/bootdev-server/internal/database"
	"database/sql"
)



type apiConfig struct {
	fileserverHits atomic.Int32
	DBQueries *database.Queries
	Platform  string
	jwtSecret string
	polkaKey  string
}


func main() {

	godotenv.Load()
	platform := os.Getenv("PLATFORM")

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("Failed to start db")
		return
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}

	dbQueries := database.New(db)


	const port = "8080"


	var apiCfg apiConfig
	apiCfg.DBQueries = dbQueries
	apiCfg.Platform = platform
	apiCfg.polkaKey = polkaKey

	// create a new http.ServeMux
	serveMultiplexer := http.NewServeMux()

	serveMultiplexer.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMultiplexer.HandleFunc("GET /api/healthz", healthHandler)
	serveMultiplexer.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	serveMultiplexer.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)
	//serveMultiplexer.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	serveMultiplexer.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	serveMultiplexer.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	serveMultiplexer.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	serveMultiplexer.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpHandler)
	serveMultiplexer.HandleFunc("POST /api/login", apiCfg.loginHandler)
	serveMultiplexer.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	serveMultiplexer.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	serveMultiplexer.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	serveMultiplexer.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	serveMultiplexer.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerWebhook)



	// create a new http server struct

	server := &http.Server{
		Addr: ":" + port,
		Handler: serveMultiplexer,
	}

	err = server.ListenAndServe()

	fmt.Printf("Serving on port: %s", port)

	if err != nil {
		fmt.Errorf("Failed to listen and serve")
		return
	}


	log.Fatal(server.ListenAndServe())
}


