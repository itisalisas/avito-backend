package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/handlers"
	middleware2 "github.com/itisalisas/avito-backend/internal/middleware"
	"github.com/itisalisas/avito-backend/internal/service/auth"
	"github.com/itisalisas/avito-backend/internal/service/product"
	"github.com/itisalisas/avito-backend/internal/service/pvz"
	"github.com/itisalisas/avito-backend/internal/service/reception"
	"github.com/itisalisas/avito-backend/internal/storage"
	my_grpc "github.com/itisalisas/avito-backend/internal/transport/grpc"
	"github.com/itisalisas/avito-backend/pkg/metrics"
	middleware3 "github.com/itisalisas/avito-backend/pkg/middleware"
)

func initializeDatabase() (*storage.DB, error) {
	db, err := storage.NewPostgres(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	if err != nil {
		return nil, err
	}
	err = db.Migrate("file://migrations/")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupRouter(authHandler *handlers.AuthHandler, pvzHandler *handlers.PvzHandler,
	productHandler *handlers.ProductHandler, receptionHandler *handlers.ReceptionHandler) http.Handler {

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

	return m
}

func startPrometheusServer() {
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
}

func RunServer() error {
	db, err := initializeDatabase()
	if err != nil {
		return err
	}
	defer func(db *storage.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close database connection: %v", err)
		}
	}(db)

	userRepo := storage.NewUserRepository(db.DB)
	pvzRepo := storage.NewPvzRepository(db.DB)
	productRepo := storage.NewProductRepository(db.DB)
	receptionRepo := storage.NewReceptionRepository(db.DB)

	authService := auth.NewAuthService(userRepo)
	pvzService := pvz.NewPvzService(pvzRepo)
	productService := product.NewProductService(productRepo, receptionRepo)
	receptionService := reception.NewReceptionService(receptionRepo)

	authHandler := handlers.NewAuthHandler(authService)
	pvzHandler := handlers.NewPvzHandler(pvzService)
	productHandler := handlers.NewProductHandler(productService)
	receptionHandler := handlers.NewReceptionHandler(receptionService)

	m := setupRouter(authHandler, pvzHandler, productHandler, receptionHandler)

	go func() {
		lis, err := net.Listen("tcp", ":3000")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		my_grpc.RegisterGRPCServer(s, pvzService)
		reflection.Register(s)

		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return http.ListenAndServe("localhost:"+os.Getenv("PORT"), m)
}

func main() {

	startPrometheusServer()

	if err := RunServer(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
