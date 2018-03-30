// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	//"github.com/jessfraz/paws/totessafe/reflector"
	//"github.com/jessfraz/paws/totessafe/reflector"
	"github.com/jessfraz/paws/totessafe/reflector"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var (
	internalPort int
	externalPort int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "totessafe",
	Short: "The totes safe binary is a security binary offered from your internal security team.",
	Long:  `Nothing to see here folks.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Here we need to run two gRPC servers
		// one for the internal server and one
		// for the external server
		ich := reflector.ConcurrentInternalListenAndServe(internalPort)
		ech := reflector.ConcurrentExternalListenAndServe(externalPort)
		for {
			select {
			case ierr := <-ich:
				logger.Critical(ierr.Error())
			case eerr := <-ech:
				logger.Critical(eerr.Error())
			}
		}
		os.Exit(0)
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
	rootCmd.Flags().IntVarP(&internalPort, "internal-port", "i", 14410, "The port for the internal gRPC server to listen on")
	rootCmd.Flags().IntVarP(&internalPort, "external-port", "e", 14411, "The port for the external gRPC server to listen on")
}
