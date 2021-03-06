/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	c "github.com/codenotary/immudb/cmd"
	"github.com/codenotary/immudb/pkg/auth"
	"github.com/spf13/cobra"
)

func checkFirstAdminLogin(err error) {
	if errMsg, matches := (&auth.ErrFirstAdminLogin{}).With("", "").Matches(err); matches {
		c.QuitToStdErr(
			fmt.Errorf(
				"===============\n"+
					"This looks like the very first admin login attempt, hence the following "+
					"credentials have been generated:%s"+
					"\nIMPORTANT: This is the only time they are shown, so make sure you remember them."+
					"\n===============", errMsg),
		)
	}
}

func (cl *commandline) login(cmd *cobra.Command) {
	ccmd := &cobra.Command{
		Use:               "login username (you will be prompted for password)",
		Short:             fmt.Sprintf("Login using the specified username and password (admin username is %s)", auth.AdminUsername),
		Aliases:           []string{"l"},
		PersistentPreRunE: cl.connect,
		PersistentPostRun: cl.disconnect,
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenFileName := cl.immuClient.GetOptions().TokenFileName
			ctx := cl.context
			user := args[0]
			if user == auth.AdminUsername {
				if _, err := os.Stat(tokenFileName); err != nil && os.IsNotExist(err) {
					if _, err := cl.immuClient.Login(ctx, []byte(user), []byte{}); err != nil {
						checkFirstAdminLogin(err)
					}
				}
			}
			pass, err := cl.passwordReader.Read("Password:")
			if err != nil {
				c.QuitToStdErr(err)
			}
			response, err := cl.immuClient.Login(ctx, []byte(user), pass)
			if err != nil {
				checkFirstAdminLogin(err)
				c.QuitWithUserError(err)
			}
			if err := ioutil.WriteFile(tokenFileName, response.Token, 0644); err != nil {
				c.QuitToStdErr(err)
			}
			fmt.Printf("logged in\n")
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.AddCommand(ccmd)
}

func (cl *commandline) logout(cmd *cobra.Command) {
	ccmd := &cobra.Command{
		Use:               "logout",
		Aliases:           []string{"x"},
		PersistentPreRunE: cl.connect,
		PersistentPostRun: cl.disconnect,
		RunE: func(cmd *cobra.Command, args []string) error {
			os.Remove(cl.immuClient.GetOptions().TokenFileName)
			fmt.Println("logged out")
			return nil
		},
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(ccmd)
}
