package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/handlers"
	middleware2 "github.com/itisalisas/avito-backend/internal/middleware"
	"github.com/itisalisas/avito-backend/internal/storage"
	"github.com/itisalisas/avito-backend/pkg/metrics"
	middleware3 "github.com/itisalisas/avito-backend/pkg/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func RunServer() {
	db, err := storage.NewPostgres()
	if err != nil {
		panic(err)
	}
	defer func(db *storage.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close database connection: %v", err)
		}
	}(db)
	err = db.Migrate()
	if err != nil {
		panic(err)
	}

	authHandler := handlers.NewAuthHandler(db.DB)
	pvzHandler := handlers.NewPvzHandler(db.DB)
	productHandler := handlers.NewProductHandler(db.DB)
	receptionHandler := handlers.NewReceptionHandler(db.DB)

	m := chi.NewRouter()
	m.Use(middleware3.MetricsMiddleware)
	m.Use(middleware.Logger)

	m.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)
	m.HandleFunc("POST /register", authHandler.Register)
	m.HandleFunc("POST /login", authHandler.Login)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Moderator)).HandleFunc("POST /pvz", pvzHandler.AddPvz)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Moderator, dto.Employee)).HandleFunc("GET /pvz", pvzHandler.GetPvz)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Employee)).HandleFunc("POST /pvz/{pvzId}/close_last_reception", receptionHandler.CloseLastReception)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Employee)).HandleFunc("POST /pvz/{pvzId}/delete_last_product", productHandler.DeleteLastProduct)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Employee)).HandleFunc("POST /receptions", receptionHandler.AddReception)
	m.With(middleware2.CheckAuth(), middleware2.CheckRole(dto.Employee)).HandleFunc("POST /products", productHandler.AddProduct)

	err = http.ListenAndServe("localhost:"+os.Getenv("PORT"), m)
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	prometheusHandler := promhttp.HandlerFor(
		metrics.Registry,
		promhttp.HandlerOpts{},
	)

	http.Handle("/metrics", prometheusHandler)

	go func() {
		if err := http.ListenAndServe(":9000", nil); err != nil {
			log.Fatalf("Prometheus server failed: %v", err)
		}
	}()

	RunServer()
}
