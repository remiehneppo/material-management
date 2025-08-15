/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"baliance.com/gooxml/document"
	"github.com/spf13/cobra"
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
		docxPath := "test-data/02M4.docx"

		doc, err := document.Open(docxPath)
		if err != nil {
			log.Fatal(err)
		}
		paras := doc.Paragraphs()

		for _, para := range paras {
			for _, run := range para.Runs() {
				if strings.Contains(run.Text(), "{project}") {
					run.ClearContent()
					run.AddText("Tàu 123")
				}
				if strings.Contains(run.Text(), "{workshop}") {
					run.ClearContent()
					run.AddText("X. Van ống")
				}
				if strings.Contains(run.Text(), "{team}") {
					run.ClearContent()
					run.AddText("New team")
				}
				if strings.Contains(run.Text(), "{description}") {
					run.ClearContent()
					run.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Praesent blandit tristique ultricies. Mauris bibendum neque nec mollis tempor. Ut pulvinar finibus sapien nec ullamcorper. Duis hendrerit quam vitae ligula viverra rhoncus. ")
				}

			}
		}

		tables := doc.Tables()

		numRqCell := tables[0].Rows()[0].Cells()[2]
		for _, para := range numRqCell.Paragraphs() {
			for _, run := range para.Runs() {
				run.ClearContent()
			}
		}
		numRqCell.Paragraphs()[0].Runs()[0].AddText("Số: 1/123/VK/25")

		materialTable := tables[1]

		newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
		indexRun := newRow.AddCell().AddParagraph().AddRun()
		indexRun.Properties().SetBold(true)
		indexRun.AddText("I")

		titleRun := newRow.AddCell().AddParagraph().AddRun()
		titleRun.Properties().SetBold(true)
		titleRun.AddText("Thiết bị A")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")
		newRow.AddCell().AddParagraph().AddRun().AddText("")

		for i := 1; i <= 10; i++ {
			newRow := materialTable.InsertRowBefore(materialTable.Rows()[len(materialTable.Rows())-1])
			newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("%d", i))
			newRow.AddCell().AddParagraph().AddRun().AddText(fmt.Sprintf("vật tư %d", i))
			newRow.AddCell().AddParagraph().AddRun().AddText("Cái")
			newRow.AddCell().AddParagraph().AddRun().AddText("")
			newRow.AddCell().AddParagraph().AddRun().AddText("100")
			newRow.AddCell().AddParagraph().AddRun().AddText("")
			newRow.AddCell().AddParagraph().AddRun().AddText("")
		}

		doc.SaveToFile("test-data/output.docx")

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
