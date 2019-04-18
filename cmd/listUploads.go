package cmd

import (
	"fmt"
	"fwatch/server"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// listUploadsCmd represents the listUploads command
var listUploadsCmd = &cobra.Command{
	Use:   "list-uploads [GROUP]",
	Short: "lists uploads",
	Long:  `list-uploads lists the uploads and time`,
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

		var uploads map[string]time.Time
		if len(args) == 0 {
			uploads = s.ListUploads()
		} else {
			panic("not impl.")

		}
		for path, t := range uploads {
			fmt.Println(path, t.Format(time.RFC3339))
		}
	},
}

func init() {
	rootCmd.AddCommand(listUploadsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listUploadsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listUploadsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
