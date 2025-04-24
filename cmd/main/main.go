package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/itisalisas/avito-backend/internal/handlers"
	middleware2 "github.com/itisalisas/avito-backend/internal/middleware"
	"net/http"
)

const port = "8080"

func RunServer() {
	m := chi.NewRouter()
	m.Use(middleware.Logger)
	m.Use(middleware2.CheckAuth())

	m.HandleFunc("POST /dummyLogin", handlers.DummyLoginHandler)
	m.HandleFunc("POST /register", handlers.RegisterHandler)
	m.HandleFunc("POST /login", handlers.LoginHandler)
	m.HandleFunc("POST /pvz", handlers.AddPvzHandler)
	m.HandleFunc("GET /pvz", handlers.GetPvzHandler)
	m.HandleFunc("POST /pvz/{pvzId}/close_last_reception", handlers.CloseLastReceptionHandler)
	m.HandleFunc("POST /pvz/{pvzId}/delete_last_product", handlers.DeleteLastProductHandler)
	m.HandleFunc("POST /receptions", handlers.AddReceptionHandler)
	m.HandleFunc("POST /products", handlers.AddProductHandler)

	err := http.ListenAndServe("localhost:"+port, m)
	if err != nil {
		panic(err)
	}
}

func main() {
	RunServer()
}
