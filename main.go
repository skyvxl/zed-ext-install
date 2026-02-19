package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:               "zed-ext-install",
		Short:             "CLI tool for installing Zed editor extensions",
		Version:           version,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(searchCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <extension-id> [version]",
		Short: "Install a Zed extension",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extID := args[0]

			paths, err := GetZedPaths()
			if err != nil {
				return err
			}

			fmt.Printf("Looking up extension %q...\n", extID)
			ext, err := FindExtension(extID)
			if err != nil {
				return err
			}

			// Override version if specified
			if len(args) > 1 {
				ext.Version = args[1]
			}

			fmt.Printf("Installing %s v%s...\n", ext.Name, ext.Version)

			if err := InstallExtension(ext, paths); err != nil {
				return err
			}

			// Update index
			idx, err := LoadIndex(paths)
			if err != nil {
				return fmt.Errorf("load index: %w", err)
			}

			if err := UpdateIndexForExtension(extID, paths, idx); err != nil {
				fmt.Printf("  warning: could not update index: %v\n", err)
			} else {
				if err := SaveIndex(paths, idx); err != nil {
					fmt.Printf("  warning: could not save index: %v\n", err)
				}
			}

			fmt.Printf("Successfully installed %s v%s\n", ext.Name, ext.Version)
			return nil
		},
	}
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <extension-id>",
		Short: "Remove an installed Zed extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extID := args[0]

			paths, err := GetZedPaths()
			if err != nil {
				return err
			}

			fmt.Printf("Removing extension %q...\n", extID)

			if err := RemoveExtension(extID, paths); err != nil {
				return err
			}

			// Update index
			idx, err := LoadIndex(paths)
			if err != nil {
				return fmt.Errorf("load index: %w", err)
			}
			RemoveFromIndex(extID, idx)
			if err := SaveIndex(paths, idx); err != nil {
				fmt.Printf("  warning: could not save index: %v\n", err)
			}

			fmt.Printf("Successfully removed %s\n", extID)
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed Zed extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := GetZedPaths()
			if err != nil {
				return err
			}

			idx, err := LoadIndex(paths)
			if err != nil {
				return err
			}

			if len(idx.Extensions) == 0 {
				fmt.Println("No extensions installed (via index.json).")

				// Also check filesystem
				entries, err := os.ReadDir(paths.Installed)
				if err == nil && len(entries) > 0 {
					fmt.Println("\nDirectories found in installed/:")
					for _, e := range entries {
						if e.IsDir() {
							fmt.Printf("  %s\n", e.Name())
						}
					}
				}
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tVERSION\tDEV")
			for id, entry := range idx.Extensions {
				fmt.Fprintf(w, "%s\t%s\t%s\t%v\n",
					id, entry.Manifest.Name, entry.Manifest.Version, entry.Dev)
			}
			w.Flush()

			return nil
		},
	}
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for Zed extensions in the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			fmt.Printf("Searching for %q...\n\n", query)
			results, err := SearchExtensions(query)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				fmt.Println("No extensions found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tVERSION\tDOWNLOADS\tDESCRIPTION")
			for _, ext := range results {
				desc := ext.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
					ext.ID, ext.Name, ext.Version, ext.DownloadCount, desc)
			}
			w.Flush()

			return nil
		},
	}
}
