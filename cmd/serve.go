package cmd

import (
	"fmt"
	"log"

	"github.com/go_video_subs/config"
	"github.com/go_video_subs/internal/delivery/http/handler"
	"github.com/go_video_subs/internal/delivery/http/router"
	repoPayment "github.com/go_video_subs/internal/repository/payment"
	repoSub "github.com/go_video_subs/internal/repository/subscription"
	repoUser "github.com/go_video_subs/internal/repository/user"
	repoVideo "github.com/go_video_subs/internal/repository/video"
	ucPayment "github.com/go_video_subs/internal/usecase/payment"
	ucUser "github.com/go_video_subs/internal/usecase/user"
	ucVideo "github.com/go_video_subs/internal/usecase/video"
	"github.com/go_video_subs/pkg/database"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Bootstrap all dependencies and start the Fiber HTTP server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if port, _ := cmd.Flags().GetString("port"); port != "" {
			cfg.App.Port = port
		}

		db, err := database.NewMariaDB(cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()
		log.Println("Database connected")

		userRepo := repoUser.New(db)
		subRepo := repoSub.New(db)
		planRepo := repoSub.NewPlan(db)
		videoRepo := repoVideo.New(db)
		paymentRepo := repoPayment.New(db)

		userUC := ucUser.New(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
		videoUC := ucVideo.New(videoRepo)
		paymentUC := ucPayment.New(paymentRepo, planRepo, subRepo)

		app := router.New(&router.Handlers{
			User:      handler.NewUserHandler(userUC),
			Auth:      handler.NewAuthHandler(userUC),
			Video:     handler.NewVideoHandler(videoUC),
			Payment:   handler.NewPaymentHandler(paymentUC),
			SubRepo:   subRepo,
			JWTSecret: cfg.JWT.Secret,
		})

		addr := fmt.Sprintf(":%s", cfg.App.Port)
		log.Printf("Server starting on http://localhost%s", addr)
		return app.Listen(addr)
	},
}

func init() {
	serveCmd.Flags().String("port", "", "Port to run the server on (overrides APP_PORT env variable)")
}
