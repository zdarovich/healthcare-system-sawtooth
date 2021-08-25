package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"healthcare-system-sawtooth/client/lib"
)

var (
	version        bool
	cfgFile        string
	name           string
	debug          bool
	bootstrapAddrs []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   lib.FamilyName,
	Short: "Decentralized cloud client application",
	Long: `Healthcare is a decentralized cloud client application.
This application is a tool for store data on a network based on Hyperledger Sawtooth.`,
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Println("Healthcare (Decentralized data client system)")
			fmt.Println("Version: " + lib.FamilyVersion)
			return
		}
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	initConfig()
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, "the version of Healthcare")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file(json)")
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", GetDefaultUsername(), "the name of user")
	rootCmd.PersistentFlags().StringVarP(&lib.TPURL, "url", "u", lib.DefaultTPURL, "the hyperledger sawtooth rest api url")
	rootCmd.PersistentFlags().StringVarP(&lib.ValidatorURL, "validator", "V", lib.DefaultValidatorURL, "the hyperledger sawtooth validator tcp url")
	rootCmd.PersistentFlags().StringVarP(&lib.PrivateKeyFile, "key", "k", lib.DefaultPrivateKeyFile, "the private key file for identity")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug version")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".SeaStorage" (without extension).
		viper.AddConfigPath(lib.DefaultConfigPath)
		viper.SetConfigName(lib.DefaultConfigFilename)
		if _, err := os.Stat(path.Join(lib.DefaultConfigPath, lib.DefaultConfigFilename+".json")); os.IsNotExist(err) {
			os.MkdirAll(lib.DefaultConfigPath, 0755)
			cf, err := os.Create(path.Join(lib.DefaultConfigPath, lib.DefaultConfigFilename+".json"))
			if err != nil {
				panic(err)
			}
			_, err = cf.Write(initConfigJSON())
			if err != nil {
				panic(err)
			}
			cf.Close()
		}
	}

	viper.SetConfigType("json")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		tpURL := viper.GetString("url")
		if tpURL == "" {
			lib.DefaultTPURL = tpURL
		}
		validatorURL := viper.GetString("validator")
		if validatorURL == "" {
			lib.DefaultValidatorURL = validatorURL
		}
		privateKeyFile := viper.GetString("key")
		if privateKeyFile == "" {
			lib.DefaultPrivateKeyFile = privateKeyFile
		}
	}
}

// initLogger config logger
func initLogger() {
	lib.Logger = logrus.New()
	lib.Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	os.MkdirAll(lib.DefaultLogPath, 0755)
	logFile, err := os.OpenFile(path.Join(lib.DefaultLogPath, "Healthcare"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logFile, err = os.OpenFile(path.Join(lib.DefaultLogPath, "Healthcare"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	lib.Logger.SetOutput(mw)
	lib.Logger.SetLevel(logrus.DebugLevel)

}

// init config in JSON format
func initConfigJSON() []byte {
	cfg := make(map[string]interface{})
	cfg["url"] = lib.DefaultTPURL
	cfg["validator"] = lib.DefaultValidatorURL
	cfg["key"] = GetDefaultKeyFile()
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		panic(err)
	}
	return data
}

// GetDefaultUsername returns the name of current system user.
func GetDefaultUsername() string {
	u, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return u.Username
}

// GetDefaultKeyFile returns the default key file named as username
// in the default key path.
func GetDefaultKeyFile() string {
	return path.Join(lib.DefaultKeyPath, GetDefaultUsername()+".priv")
}
