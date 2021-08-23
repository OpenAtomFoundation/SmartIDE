/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os/exec"
	"bufio"
	"io"
	"github.com/spf13/cobra"
)

//default values
var SmartIDEPort=3000;
var SmartIDEImage="smartide-registry.cn-beijing.cr.aliyuncs.com/library/smartide-node:latest";
var SmartIDEName="smartide"

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "快速创建并启动SmartIDE开发环境",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("SmartIDE启动中......")
		dockerRunCommand := fmt.Sprintf(`docker run -i --user root --name=%s --init -p %d:3000 --expose 3001 -p 3001:3001 -v "$(pwd):/home/project" %s --inspect=0.0.0.0:3001`, SmartIDEName, SmartIDEPort,SmartIDEImage)
		command := exec.Command("/bin/sh","-c",dockerRunCommand)
		stdout, err := command.StdoutPipe()

		if err != nil {
			fmt.Println(err)
		}
		
		command.Start()

		reader := bufio.NewReader(stdout)

		//实时循环读取输出流中的一行内容
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			fmt.Println(line)
		}

		command.Wait()
		fmt.Println("SmartIDE启动完毕......")


	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
