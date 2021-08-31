package commands

import (
	"github.com/spf13/cobra"
	"healthcare-system-sawtooth/client/lib"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate key for identity",
	Long:  `Generate private and public key for identity of sawtooth blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {

		lib.GenerateKey(name, lib.DefaultKeyPath)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	rootCmd.PersistentFlags().StringVarP(&lib.DefaultKeyPath, "path", "p", lib.DefaultConfigPath, "the hyperledger sawtooth key path")
}
