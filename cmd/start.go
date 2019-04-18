package cmd

import (
	"fmt"
	"fwatch/server"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start watching files listed in config",
	Long:  ``,
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

		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig
			s.Stop()
		}()

		err = s.Start()

		if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
