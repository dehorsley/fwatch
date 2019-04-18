package cmd

import (
	"fmt"
	"fwatch/server"
	"os"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [GROUP]",
	Short: "list the contents of a group",
	Long:  `list the contents of a group`,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(cfgFile)
		if err != nil {
			fmt.Printf("error opening config file %s\n", cfgFile)
			os.Exit(1)
		}
		defer f.Close()
		s, err := server.New(f)
		if err != nil {
			fmt.Printf("error loading config: %s\n", err)
			os.Exit(1)
		}

		paths := make([]string, 0)
		if len(args) == 0 {
			for _, group := range s.Groups {
				paths = append(paths, group.List()...)
			}
		} else {
			for _, g := range args {
				group, ok := s.Groups[g]
				if !ok {
					fmt.Printf("%q is not a a group\n", g)
					os.Exit(1)
				}
				paths = append(paths, group.List()...)
			}

		}
		for _, path := range paths {
			fmt.Println(path)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
