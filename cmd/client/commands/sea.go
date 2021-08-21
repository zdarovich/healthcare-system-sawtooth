package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/sea"
)

// seaCmd represents the sea command
var seaCmd = &cobra.Command{
	Use:   "sea",
	Short: "SeaStorage Sea Command Client",
	Long: `SeaStorage Sea Command Client is a platform support
communicating with the transaction processor
and listening for the P2P network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			fmt.Println(errors.New("the name of user/sea is required"))
			os.Exit(0)
		}
		cli, err := sea.NewSeaClient(name, lib.PrivateKeyFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer cli.Close()
		err = cli.Sync()
		if err != nil {
			err := cli.SeaRegister()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		cli.Bootstrap(lib.PrivateKeyFile, lib.StoragePath, lib.StorageSize, lib.BootstrapAddrs)
	},
}

func init() {
	rootCmd.AddCommand(seaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// seaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// seaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	seaCmd.Flags().StringVarP(&lib.StoragePath, "path", "P", lib.DefaultStoragePath, "the path for storage")
	seaCmd.Flags().Int64VarP(&lib.StorageSize, "size", "s", lib.DefaultStorageSize, "the size for storage")
}
