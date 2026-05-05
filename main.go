package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/config"
	"github.com/milua25/e-commerce-backend/controllers"
	"github.com/milua25/e-commerce-backend/database"
	"github.com/milua25/e-commerce-backend/middlewares"
	"github.com/milua25/e-commerce-backend/routes"
	"github.com/milua25/e-commerce-backend/tokens"
)

func main() {

	// Load configuration
	cfg, missingVars := config.LoadConfig()
	// Log missing environment variables, if any
	if len(missingVars) > 0 {
		log.Printf("Warning: Missing environment variables: %v\n", missingVars)
		for _, v := range missingVars {
			log.Printf(" - %s\n", v)
		}
		os.Exit(1)
	}

	logger := &database.Logger{}
	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?retryWrites=true&w=majority",
		url.QueryEscape(cfg.DBConfig.User),
		url.QueryEscape(cfg.DBConfig.Password),
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
	)
	client, err := database.DBSet(mongoURI, logger)
	if err != nil {
		log.Fatalf("failed to initialize MongoDB: %v", err)
	}

	// Ensure MongoDB client is disconnected when the application exits
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	// Context with timeout for ensuring collections exist before starting the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure necessary collections exist before starting the server
	err = database.EnsureCollections(ctx, client, cfg.DBConfig.DBName, cfg.DBConfig.UserCollection, cfg.DBConfig.ProductCollection)
	if err != nil {
		log.Fatalf("failed to ensure MongoDB collections: %v", err)
	}

	app := controllers.NewApplication(
		database.ProductData(client, cfg.DBConfig.ProductCollection, cfg.DBConfig.DBName),
		database.UserData(client, cfg.DBConfig.UserCollection, cfg.DBConfig.DBName),
		cfg.JWTConfig.SecretKey,
	)
	authService := tokens.NewAuthService(cfg.JWTConfig.SecretKey)

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	public := router.Group("/")
	routes.UserRoutes(public, app)
	public.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Welcome to the E-commerce API",
		})
	})

	protected := router.Group("/")
	protected.Use(middlewares.Authentication(authService))
	{
		//user routes
		protected.PATCH("/users/:id/update", app.UpdateUserDetails())
		protected.GET("/users/count", app.GetUserCount())
		protected.POST("/users/:id/address/add", app.CreateAddress())
		// protected.DELETE("/user/profile/delete", app.DeleteUserProfile())

		protected.POST("/products/add", app.AddProductToDatabase())
		protected.GET("/products", app.ProductViewerAdmin())
		protected.GET("/addtocart", app.AddProductToCart())
		// protected.GET("/removeitem", app.RemoveItem())
		protected.GET("/viewcart", app.ViewCart())
		// protected.GET("/cartcheckout", app.Checkout())
		protected.GET("/instantbuy", app.InstantBuy())
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	errChan := make(chan error, 1)
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Printf("Starting server on %s...", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
		// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	case <-sigCtx.Done():
		// Restore default behavior on the interrupt signal and notify user of shutdown.
		// stop()
		log.Println("shutting down gracefully, press Ctrl+C again to force")
	}

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
