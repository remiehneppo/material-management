package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/remiehneppo/material-management/config"
	"github.com/remiehneppo/material-management/internal/database"
	"github.com/remiehneppo/material-management/internal/migration"
	"github.com/spf13/cobra"
)

var migrateDomainCmd = &cobra.Command{
	Use:   "migrate-domain",
	Short: "Preflight or apply the domain-module data migration",
	RunE: func(command *cobra.Command, _ []string) error {
		configPath, _ := command.Flags().GetString("config")
		apply, _ := command.Flags().GetBool("apply")
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(command.Context(), 2*time.Minute)
		defer cancel()
		db := database.NewMongoDatabase(cfg.MongoDB.URI, cfg.MongoDB.Database)
		if err := db.Connect(ctx); err != nil {
			return err
		}
		defer db.Disconnect(context.Background())
		runner := migration.NewRunner(db.Client(), db.DB())
		report, err := runner.Preflight(ctx)
		if err != nil {
			return err
		}
		encoded, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(encoded))
		if !apply {
			return nil
		}
		if !report.Ready() {
			return fmt.Errorf("preflight found %d conflict(s); no writes performed", len(report.Conflicts))
		}
		return runner.Apply(ctx)
	},
}

func init() {
	rootCmd.AddCommand(migrateDomainCmd)
	migrateDomainCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")
	migrateDomainCmd.Flags().Bool("apply", false, "Apply after a successful read-only preflight")
}
