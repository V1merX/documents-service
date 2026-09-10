package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/V1merX/documents-service/internal/di"
	"github.com/V1merX/documents-service/internal/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type App struct {
	httpServer  *http.Server
	initOnce    *sync.Once
	diContainer *di.Container
}

func New() (*App, error) {
	diC, err := di.NewContainer()
	if err != nil {
		return nil, err
	}

	return &App{
		initOnce:    &sync.Once{},
		diContainer: diC,
	}, nil
}

func (a *App) Run(ctx context.Context) (err error) {
	if err := a.initDeps(); err != nil {
		return err
	}

	a.diContainer.Logger().Info("Starting http server", zap.String("address", a.httpServer.Addr))

	go func() {
		if e := a.httpServer.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			err = e
			a.diContainer.Logger().Error("Failed to listen and serve", zap.Error(err))
		}

		return
	}()

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	<-gracefulStop

	a.diContainer.Logger().Info("Shutting down application...")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.diContainer.Logger().Error("Failed to shutdown http server", zap.Error(err))
	}

	a.diContainer.Flush()

	return nil
}

func (a *App) initDeps() (err error) {
	a.initOnce.Do(func() {
		deps := []func() error{
			a.initHTTPServer,
		}

		for _, dep := range deps {
			if e := dep(); e != nil {
				err = e
				return
			}
		}
	})

	return
}

func (a *App) initRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.CleanPath)
	r.Use(chiMiddleware.ClientIPFromRemoteAddr)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Route("/docs", func(r chi.Router) {
			r.Use(chiMiddleware.GetHead)

			r.With(middleware.AuthMiddleware(a.diContainer.UserService(), a.diContainer.Logger())).
				Get("/", a.diContainer.DocsHandler().List)

			r.Post("/", a.diContainer.DocsHandler().Create)

			r.With(middleware.AuthMiddleware(a.diContainer.UserService(), a.diContainer.Logger())).
				Get("/{id}", a.diContainer.DocsHandler().Get)

			r.With(middleware.AuthMiddleware(a.diContainer.UserService(), a.diContainer.Logger())).
				Delete("/{id}", a.diContainer.DocsHandler().Delete)
		})

		r.Post("/register", a.diContainer.AuthHandler().Register)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/", a.diContainer.AuthHandler().Auth)
			r.Delete("/{token}", a.diContainer.AuthHandler().Logout)
		})
	})

	return r
}

func (a *App) initHTTPServer() error {
	a.httpServer = &http.Server{
		Addr:              a.diContainer.Config().HTTPServer.Addr,
		ReadTimeout:       a.diContainer.Config().HTTPServer.ReadTimeout,
		ReadHeaderTimeout: a.diContainer.Config().HTTPServer.ReadHeaderTimeout,
		WriteTimeout:      a.diContainer.Config().HTTPServer.WriteTimeout,
		IdleTimeout:       a.diContainer.Config().HTTPServer.IdleTimeout,
		MaxHeaderBytes:    a.diContainer.Config().HTTPServer.MaxHeaderBytes,
		Handler:           a.initRouter(),
	}

	return nil
}
