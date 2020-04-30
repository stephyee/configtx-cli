/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ConfigFilePath   string
	Policy           string
	ACLs             string
	Capability       string
	OrderingEndpoint string
	OrgName          string
)

// The main command describes the service and
// defaults to printing the help message.
var mainCmd = &cobra.Command{Use: "peer"}

func main() {
	// For environment variables.
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Define command-line flags that are valid for all peer commands and
	// subcommands.
	mainFlags := mainCmd.PersistentFlags()

	mainFlags.String("logging-level", "", "Legacy logging level flag")
	viper.BindPFlag("logging_level", mainFlags.Lookup("logging-level"))
	mainFlags.MarkHidden("logging-level")

	mainCmd.AddCommand()

	// On failure Cobra prints the usage message and error string, so we only
	// need to exit with a non-0 status
	if mainCmd.Execute() != nil {
		os.Exit(1)
	}
}

// endpointsCmd represents the endpoints command
var endpointsCmd = &cobra.Command{
	Use:   "endpoints",
	Short: "Updates orderer endpoint",
	Long: `Adds an orderer's endpoint to an existing channel config transaction. If
	the same endpoint already exist in current configuration, this will be a no-op.
  For example:
configtx orderer update endpoints --orgName Org1 --endpoint 127.0.0.1:8080
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("endpoints called")
		updateEndpoints()
	},
}

func updateEndpoints() {
	config, err := readBlock(ConfigFilePath)
	if err != nil {
		panic(err)
	}

	if len(strings.Split(OrderingEndpoint, ":")) != 2 {
		panic("ordering service endpoint %s is not valid or missing")
	}

	endpoint := strings.Split(OrderingEndpoint, ":")

	port, err := strconv.Atoi(endpoint[1])
	if err != nil {
		panic(err)
	}

	address := configtx.Address{
		Host: endpoint[0],
		Port: port,
	}

	err = config.SetOrdererEndpoint(OrgName, address)
	if err != nil {
		panic(err)
	}
}

func readBlock(cfgPath string) (configtx.ConfigTx, error) {
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return configtx.ConfigTx{}, fmt.Errorf("could not read block from file %s", cfgPath)
	}

	blk, err := protoutil.UnmarshalBlock(data)
	if err != nil {
		panic(err)
	}

	payload, err := protoutil.UnmarshalPayload(blk.Data.Data[0])
	if err != nil {
		panic(err)
	}

	config := &cb.Config{}

	err = proto.Unmarshal(payload.Data, config)
	if err != nil {
		panic(err)
	}

	return configtx.New(config), nil
}
