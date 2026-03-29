package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/cgorricho/internet-monitor/internal/config"
	"github.com/cgorricho/internet-monitor/internal/database"
	"github.com/cgorricho/internet-monitor/internal/dashboard"
	"github.com/cgorricho/internet-monitor/internal/monitor"
	"github.com/cgorricho/internet-monitor/internal/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "internet-monitor",
		Short: "Internet Connection Performance Monitor",
		Long: `A comprehensive internet connection monitoring tool that tracks network performance
and generates comparative reports across multiple machines.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Initialize command
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize database and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.New(cfg.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			if err := db.Migrate(); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			fmt.Println("✅ Database initialized successfully")
			return nil
		},
	}

	// Monitor command
	var monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Start network monitoring service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.New(cfg.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			mon := monitor.New(cfg, db)
			
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle graceful shutdown
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-c
				fmt.Println("\n🛑 Shutting down monitor...")
				cancel()
			}()

			fmt.Println("🚀 Starting internet monitor...")
			return mon.Start(ctx)
		},
	}

	// Dashboard command
	var dashboardCmd = &cobra.Command{
		Use:   "dashboard",
		Short: "Generate static HTML dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.New(cfg.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			dash := dashboard.New(cfg, db)
			
			hours, _ := cmd.Flags().GetInt("hours")
			compare, _ := cmd.Flags().GetBool("compare")
			output, _ := cmd.Flags().GetString("output")
			noBrowser, _ := cmd.Flags().GetBool("no-browser")

			opts := dashboard.GenerateOptions{
				Hours:     hours,
				Compare:   compare,
				Output:    output,
				NoBrowser: noBrowser,
			}

			return dash.Generate(opts)
		},
	}

	// Add dashboard flags
	dashboardCmd.Flags().IntP("hours", "h", 24, "Hours of data to include")
	dashboardCmd.Flags().BoolP("compare", "c", false, "Generate comparative dashboard")
	dashboardCmd.Flags().StringP("output", "o", "dashboard.html", "Output filename")
	dashboardCmd.Flags().Bool("no-browser", false, "Don't open browser automatically")

	// Server command (for pairing API)
	var serverCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start API server for pairing",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.New(cfg.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			srv := server.New(cfg, db)
			
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle graceful shutdown
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-c
				fmt.Println("\n🛑 Shutting down server...")
				cancel()
			}()

			fmt.Printf("🌐 Starting API server on %s:%d...\n", cfg.Server.Host, cfg.Server.Port)
			return srv.Start(ctx)
		},
	}

	// Pair command
	var pairCmd = &cobra.Command{
		Use:   "pair [code]",
		Short: "Pair with another machine",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if len(args) == 0 {
				// Generate pairing code
				fmt.Println("🔗 Generating pairing code...")
				fmt.Println("Feature coming soon!")
				return nil
			}

			// Join with pairing code
			code := args[0]
			fmt.Printf("🤝 Pairing with code: %s\n", code)
			fmt.Println("Feature coming soon!")
			return nil
		},
	}

	// Status command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			db, err := database.New(cfg.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			stats, err := db.GetStats()
			if err != nil {
				return fmt.Errorf("failed to get database stats: %w", err)
			}

			fmt.Printf("📊 Internet Monitor Status\n")
			fmt.Printf("Database: %s\n", cfg.Database.Path)
			fmt.Printf("Total measurements: %d\n", stats.MeasurementCount)
			fmt.Printf("Database size: %.2f MB\n", stats.DatabaseSizeMB)
			
			if stats.LastMeasurement != nil {
				fmt.Printf("Last measurement: %s\n", stats.LastMeasurement.Timestamp.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(pairCmd)
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}