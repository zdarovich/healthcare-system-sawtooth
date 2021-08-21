package commands

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"healthcare-system-sawtooth/client/lib"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate key for identity",
	Long:  `Generate private and public key for identity of sawtooth blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		u, err := user.Current()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pathPrompt := &promptui.Prompt{
			Label:     "Key Path (default: $HOME/.SeaStorage/keys) ",
			Templates: commandTemplates,
			Default:   path.Join(u.HomeDir, ".SeaStorage", "keys"),
		}
		keyFilePath, err := pathPrompt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		namePrompt := &promptui.Prompt{
			Label:     "Key name (default: $USERNAME) ",
			Templates: commandTemplates,
			Default:   u.Username,
		}
		keyFileName, err := namePrompt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
		confirmPrompt := &promptui.Prompt{
			Label:     "Key file exists, overwrite? [Y/n]",
			Templates: commandTemplates,
			Default:   "y",
		}
		if _, err = os.Stat(path.Join(keyFilePath, keyFileName+".priv")); err == nil {
			result, err := confirmPrompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			switch result {
			case "y", "Y":
			default:
				return
			}
		} else if _, err = os.Stat(path.Join(keyFilePath, keyFileName+".pub")); err == nil {
			result, err := confirmPrompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			switch result {
			case "y", "Y":
			default:
				return
			}
		}
		lib.GenerateKey(keyFileName, keyFilePath)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
