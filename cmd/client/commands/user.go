package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"healthcare-system-sawtooth/client/db/models"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	tpStorage "healthcare-system-sawtooth/tp/storage"
	"os"
	"strings"
)

var userCommands = []string{
	"register",
	"sync",
	"whoami",
	"create",
	"share",
	"ls",
	"ls-users",
	"ls-shared",
	"get",
	"get-shared",
	"request-as-third-party",
	"request-as-trusted-party",
	"list-requests",
	"process-request",
	"batch-upload",
	"exit",
}

var (
	errMissingOperand = errors.New("missing operand")
	errInvalidPath    = errors.New("invalid path")
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Healthcare User Command Client",
	Long: `Healthcare User Command Client is a platform support
communicating with the transaction processor.`,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			fmt.Println(errors.New("the name of user is required"))
			os.Exit(0)
		}
		cli, err := user.NewUserClient(name, lib.PrivateKeyFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if cli.User != nil {
			fmt.Println("Already register.")
		} else {
			err = cli.UserRegister()
			if err != nil {
				fmt.Println(err)
			}
		}
		defer cli.Close()
		for {
			prompt := promptui.Prompt{
				Label:     name + " ",
				Templates: commandTemplates,
				Validate: func(s string) error {
					commands := strings.Fields(s)
					if len(commands) == 0 {
						return nil
					}
					for _, c := range userCommands {
						if c == commands[0] {
							return nil
						}
					}
					return fmt.Errorf("command not found: %v", commands[0])
				},
			}
			err = nil
			input, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				return
			}
			commands := strings.Fields(input)
			if len(commands) == 0 {
				continue
			}
			if commands[0] == "exit" {
				os.Exit(1)
				return
			} else if commands[0] == "register" {
				if cli.User != nil {
					fmt.Println("Already register.")
					continue
				}
				err = cli.UserRegister()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("User register success.")
				}
				continue
			} else if cli.User == nil {
				fmt.Println("need register firstly")
				continue
			}
			switch commands[0] {
			case "sync":
				err = cli.Sync()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("sync success")
				}
			case "whoami":
				cli.ClientFramework.Whoami()
			case "ls-users":
				if len(commands) == 1 {
					err := cli.ListUsers()
					if err != nil {
						fmt.Println(err)
						return
					}
					for addr, u := range cli.QueryCache {
						fmt.Printf("Address: %s User: %s \n", addr, u.Name)
					}
				} else {
					fmt.Println(errMissingOperand)
				}
			case "create":
				if len(commands) < 3 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 3 {
					fmt.Println(errInvalidPath)
				} else {
					_, err = cli.CreatePatientData(commands[1], commands[2], 0)
					if err != nil {
						fmt.Println(err)
					}
				}
			case "share":
				if len(commands) < 3 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 3 {
					fmt.Println(errInvalidPath)
				} else {
					err = cli.ShareData(commands[1], commands[2])
					if err != nil {
						fmt.Println(err)
					}
				}
			case "get":
				if len(commands) < 2 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 2 {
					fmt.Println(errInvalidPath)
				} else {
					_, data, err := cli.GetPatientData(commands[1])
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(data)
					}
				}
			case "ls":
				iNodes, err := cli.ListPatientData()
				if err != nil {
					fmt.Println(err)
				} else {
					for _, n := range iNodes {
						printINode(n)
					}
				}

			case "get-shared":
				if len(commands) < 3 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 3 {
					fmt.Println(errInvalidPath)
				} else {
					_, data, err := cli.GetSharedPatientData(commands[1], commands[2])
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(data)
					}
				}
			case "ls-shared":
				if len(commands) < 2 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 2 {
					fmt.Println(errInvalidPath)
				} else {
					iNodes, err := cli.ListSharedPatientData(commands[1])
					if err != nil {
						fmt.Println(err)
					} else {
						for _, n := range iNodes {
							printINode(n)
						}
					}
				}
			case "request-as-third-party":
				if len(commands) < 4 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 4 {
					fmt.Println(errInvalidPath)
				} else {
					err := cli.RequestData(commands[1], commands[2], commands[3])
					if err != nil {
						fmt.Println(err)
					}
				}
			case "request-as-trusted-party":
				if len(commands) < 2 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 2 {
					fmt.Println(errInvalidPath)
				} else {
					err := cli.RequestData(commands[1], commands[1], "0")
					if err != nil {
						fmt.Println(err)
					}
				}
			case "list-requests":
				reqs, err := cli.ListRequests()
				if err != nil {
					fmt.Println(err)
				} else {
					for _, n := range reqs {
						printRequest(n)
					}
				}
			case "process-request":
				if len(commands) < 3 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 3 {
					fmt.Println(errInvalidPath)
				} else {
					var accept bool
					if commands[2] == "true" {
						accept = true
					}
					err := cli.ProcessRequest(commands[1], accept)
					if err != nil {
						fmt.Println(err)
					}
				}
			case "batch-upload":
				if len(commands) < 2 {
					fmt.Println(errMissingOperand)
				} else if len(commands) > 2 {
					fmt.Println(errInvalidPath)
				} else {

					err := cli.BatchUpload(commands[1])
					if err != nil {
						fmt.Println(err)
					}
				}
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
}

// printINode display the information of iNode.
func printINode(iNode tpStorage.INode) {
	data, err := json.MarshalIndent(iNode, "", "\t")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(data))
	}
}

func printRequest(req *models.Request) {
	data, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(data))
	}
}
