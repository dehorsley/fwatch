package cmd

import (
	"fmt"
	"fwatch/server"
	"os"

	"github.com/spf13/cobra"
)

// ignoreCmd represents the ignore command
var ignoreCmd = &cobra.Command{
	Use:   "ignore [GROUP|FILE]",
	Short: "ignore most recent changes",
	Long: `ignore sets the sync time in the database to now this has the effect
of preventing the file or group from being uploaded until it changes
again.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("specify a group or file to ignore")
			os.Exit(1)
		}

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

		var paths []string

		if group, ok := s.Groups[args[0]]; ok {
			paths = group.List()
		} else {
			// TODO: check if file is in a group
			path := args[0]
			_, err := os.Stat(path)
			if err != nil {
				fmt.Printf("unknown file or group %q\n", path)
				os.Exit(1)
			}
			paths = []string{args[0]}
		}

		s.Ignore(paths...)
	},
}

func init() {
	rootCmd.AddCommand(ignoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ignoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ignoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
