package cmd

import (
	"fmt"
	"fwatch/server"
	"os"

	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var reuploadCmd = &cobra.Command{
	Use:   "reupload [GROUP|FILE]",
	Short: "reupload a group or a file",
	Long: `reupload deletes all upload records of a group in the upload database,
marking them for re-upload when the fwatch is next run.
Eg
    reupload vgosDB
    reupload /shared/gemini/500/solve/apriori_files/blokq.dat
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("specify a file or group re-upload")
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

		s.Reset(paths...)
	},
}

func init() {
	rootCmd.AddCommand(reuploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clearCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clearCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
