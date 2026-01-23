package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var genDocCmd = &cobra.Command{
	Use:    "gen-doc",
	Short:  "Generate Markdown documentation for the CLI",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("dir")
		if outputDir == "" {
			return errors.New("output directory is required")
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// Generate standard markdown docs
		// We use a custom file prepender to add Doks-compatible frontmatter
		err := doc.GenMarkdownTreeCustom(rootCmd, outputDir, filePrepender, linkHandler)
		if err != nil {
			return err
		}

		fmt.Printf("Documentation generated in %s\n", outputDir)
		return nil
	},
}

func init() {
	genDocCmd.Flags().StringP("dir", "d", "", "Output directory for documentation")
	rootCmd.AddCommand(genDocCmd)
}

func filePrepender(filename string) string {
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, filepath.Ext(name))
	// Convert aix_mcp_add.md -> mcp add
	title := strings.ReplaceAll(base, "_", " ")

	return fmt.Sprintf(`---
title: "%s"
description: "Reference for %s command"
draft: false
toc: true
---
`, title, title)
}

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	return "/docs/reference/" + strings.ToLower(base) + "/"
}
