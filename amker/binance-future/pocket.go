package main

import (
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/migrations/logs"
	"github.com/pocketbase/pocketbase/tools/migrate"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type migrationsConnection struct {
	DB             *dbx.DB
	MigrationsList migrate.MigrationsList
}

func migrationsConnectionsMap(app core.App) map[string]migrationsConnection {
	return map[string]migrationsConnection{
		"db": {
			DB:             app.DB(),
			MigrationsList: migrations.AppMigrations,
		},
		"logs": {
			DB:             app.LogsDB(),
			MigrationsList: logs.LogsMigrations,
		},
	}
}
func runMigrations(app core.App) error {
	connections := migrationsConnectionsMap(app)

	for _, c := range connections {
		//fmt.Println("--------", k, c.DB)
		runner, err := migrate.NewRunner(c.DB, c.MigrationsList)
		if err != nil {
			return err
		}

		if _, err := runner.Up(); err != nil {
			return err
		}
	}

	return nil
}
func main() {
	app := core.NewBaseApp("tempDir", "pb_test_env", false)

	// load data dir and db connections
	if err := app.Bootstrap(); err != nil {

	}

	var allowedOrigins []string
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// add new "GET /api/hello" route to the app router (echo)
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/hello",
			Handler: func(c echo.Context) error {
				return c.String(200, "Hello world!")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
			},
		})

		return nil
	})

	// ensure that the latest migrations are applied before starting the server

	fmt.Println("4444")

	if err := runMigrations(app); err != nil {
		fmt.Println(err, "3333333333")
		panic(err)
	}
	fmt.Println("5555")

	// reload app settings in case a new default value was set with a migration
	// (or if this is the first time the init migration was executed)
	if err := app.RefreshSettings(); err != nil {
		color.Yellow("=====================================")
		color.Yellow("WARNING - Settings load error! \n%v", err)
		color.Yellow("Fallback to the application defaults.")
		color.Yellow("=====================================")
	}

	router, err := apis.InitApi(app)
	if err != nil {
		panic(err)
	}
	allowedOrigins = []string{"*"}
	// configure cors
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// start http server
	// ---
	mainAddr := "127.0.0.1:8090"

	mainHost, _, _ := net.SplitHostPort(mainAddr)

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(filepath.Join(app.DataDir(), ".autocert_cache")),
		HostPolicy: autocert.HostWhitelist(mainHost, "www."+mainHost),
	}

	serverConfig := &http.Server{
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			NextProtos:     []string{acme.ALPNProto},
		},
		ReadTimeout: 60 * time.Second,
		// WriteTimeout: 60 * time.Second, // breaks sse!
		Handler: router,
		Addr:    mainAddr,
	}
	showStartBanner := true
	if showStartBanner {
		schema := "http"

		regular := color.New()
		bold := color.New(color.Bold).Add(color.FgGreen)
		bold.Printf("> Server started at: %s\n", color.CyanString("%s://%s", schema, serverConfig.Addr))
		regular.Printf("  - REST API: %s\n", color.CyanString("%s://%s/api/", schema, serverConfig.Addr))
		regular.Printf("  - Admin UI: %s\n", color.CyanString("%s://%s/_/", schema, serverConfig.Addr))
	}

	var serveErr error
	serveErr = serverConfig.ListenAndServe()

	if serveErr != http.ErrServerClosed {
		log.Fatalln(serveErr)
	}
}
