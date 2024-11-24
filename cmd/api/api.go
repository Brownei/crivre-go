package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/brownei/crivre-go/store"
	"github.com/brownei/crivre-go/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type application struct {
	addr   string
	logger *zap.SugaredLogger
	store  *store.Store
}

func NewApplication(logger *zap.SugaredLogger, store *store.Store) *application {
	return &application{
		addr:   ":8000",
		logger: logger,
		store:  store,
	}
}

func (a *application) Run() error {
	r := chi.NewRouter()
	server := &http.Server{
		Addr:    a.addr,
		Handler: r,
	}

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, "Welcome To Crivre!")
	})
	r.Route("/auth", a.AllAuthRoutes)

	// Run the server in a goroutine so it doesn't block
	go func() {
		log.Printf("Running currently on %s", a.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Set up channel on which to send signal notifications.
	// We’ll accept graceful shutdowns when quit via SIGINT (Ctrl+C) or SIGTERM.
	// SIGKILL, SIGQUIT will not be caught.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	// Block until we receive a signal.
	<-stop

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown.
	log.Println("Shutting down server gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	return nil
}
