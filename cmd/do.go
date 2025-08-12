/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

// doCmd represents the do command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("do called")

		f, err := excelize.OpenFile("test-data/ex1.xlsx")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			// Close the spreadsheet
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		rows, err := f.GetRows("Sheet1")
		if err != nil {
			fmt.Println(err)
			return
		}
		// Iterate over the rows and print cell values
		// rowTypes := []string{"Equipment", "Consumable", "Replacement", "Material"}
		materialType := ""
		for _, row := range rows[1:] { // Skip header row

			cell0 := strings.TrimSpace(strings.ToLower(row[0]))
			cell1 := strings.TrimSpace(strings.ToLower(row[1]))

			if cell0 != "" && cell0 != "-" && cell0 != "*" {
				materialType = ""
				fmt.Printf("Hạng mục: %s\n", cell1)
			}

			if strings.Contains(cell1, "vật tư thay thế") {
				materialType = "replacement"
			} else if strings.Contains(cell1, "vật tư tiêu hao") {
				materialType = "consumable"
			}

			if materialType == "replacement" && cell0 == "-" {
				cell2 := strings.TrimSpace(strings.ToLower(row[2]))
				cell3 := strings.TrimSpace(strings.ToLower(row[3]))
				fmt.Printf("Vật tư thay thế: %s: %s %s\n", cell1, cell3, cell2)
			}
			if materialType == "consumable" && cell0 == "-" {
				cell2 := strings.TrimSpace(strings.ToLower(row[2]))
				cell3 := strings.TrimSpace(strings.ToLower(row[3]))
				fmt.Printf("Vật tư tiêu hao: %s: %s %s\n", cell1, cell3, cell2)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(doCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
