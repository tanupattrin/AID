// Copyright 2019 The CVPM Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package CVPM (main) implements the CLI for CVPM.
To get started, use:

cvpm help

to get a detailed explanation.
*/
package main

import (
	"log"
	"os"

	"github.com/getsentry/raven-go"
	"github.com/unarxiv/cvpm/pkg/config"
	"github.com/urfave/cli"
)

var (
	Version string
	Build   string
)

func main() {
	validateConfig()
	// sessionToken := getCache("session-token")
	var currentUser User
	// if sessionToken != "" {
	//		currentUser = User{"", "", sessionToken}
	// currentUser.become()
	// }
	cvpm := cli.NewApp()
	cvpm.Name = "CVPM"
	cvpm.Usage = "Computer Vision Package Manager"
	cvpm.Version = "0.0.3@alpha"
	cvpm.Commands = []cli.Command{
		{
			Name: "login",
			Action: func(c *cli.Context) error {
				LoginHandler(c)
				return nil
			},
		},
		{
			Name: "install",
			Action: func(c *cli.Context) error {
				InstallHandler(c)
				return nil
			},
		},
		{
			Name: "list",
			Action: func(c *cli.Context) error {
				listRepos(c)
				return nil
			},
		},
		{
			Name: "daemon",
			Action: func(c *cli.Context) error {
				DaemonHandler(c)
				return nil
			},
		},
		{
			Name: "repo",
			Action: func(c *cli.Context) error {
				RepoHandler(c)
				return nil
			},
		},
		{
			Name: "config",
			Action: func(c *cli.Context) error {
				ConfigHandler(c)
				return nil
			},
		},
	}
	err := cvpmapp.Run(os.Args)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatal(err)
	}
}
