/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"ccspellcheck/bloomfilter"
	"os"

	"github.com/spf13/cobra"
)

var BuildFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ccspellcheck",
	Short: "Spell Checker Using A Bloom Filter",
	Long: `
	Spell checker that can determine if a word is probably spelt correctly without having to store the full list of words.
	Thus the spell checker can use less storage (disk or memory). A task that is much less relevant these days, 
	but 20 years ago was incredibly useful on low storage devices.
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		outputFile := "data/word.bf"
		if BuildFlag {
			bloomfilter.BuildBloomFilter(inputFile, outputFile, 0.15)
		} else {
			bloomfilter.SpellCheck(outputFile, args)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().BoolVar(&BuildFlag, "build", false, "Build bloom filter")
	rootCmd.PersistentFlags().BoolVarP(&BuildFlag, "build", "b", false, "Build bloom filter")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
