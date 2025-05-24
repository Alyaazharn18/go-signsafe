package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ewallet-backend/auth"
	"ewallet-backend/db"
	"ewallet-backend/handlers"
	"ewallet-backend/utils"

	"github.com/gorilla/mux"
)

func main() {
	db.InitRedis()
	db.InitPostgres()

	r := mux.NewRouter()

	mainLogger := log.New(os.Stdout, "[MAIN] ", log.LstdFlags)

	r.Use(auth.CorsMiddleware)

	r.Handle("/api/auth/register", http.HandlerFunc(handlers.RegisterHandler)).Methods("POST", "OPTIONS")

	r.Handle("/api/transfer", auth.SignSafeMiddleware(http.HandlerFunc(handlers.TransferHandler))).Methods("POST", "OPTIONS")

	r.Handle("/api/history", auth.SignSafeMiddleware(http.HandlerFunc(handlers.HistoryHandler))).Methods("GET", "OPTIONS")

	r.Handle("/api/topup", auth.SignSafeMiddleware(http.HandlerFunc(handlers.TopupHandler))).Methods("POST", "OPTIONS")

	r.Handle("/api/users", auth.SignSafeMiddleware(http.HandlerFunc(handlers.UserHandler))).Methods("GET", "OPTIONS")

	r.Handle("/api/users/{id}", auth.SignSafeMiddleware(http.HandlerFunc(handlers.SingleUserHandler))).Methods("GET", "OPTIONS")

	mainLogger.Println("Router configuration complete.")
	mainLogger.Println("Routes configured:")
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		mainLogger.Printf("ROUTE: %s Methods: %v", pathTemplate, methods)
		return nil
	})

	PORT := utils.GetEnv("PORT", "8000")

	mainLogger.Printf("Server running on port %s...", PORT)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", PORT), r); err != nil {
		mainLogger.Fatal("Server failed to start: ", err)
	}
}
