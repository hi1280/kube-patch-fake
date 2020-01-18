/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unsafe"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

var namespace string
var kubeconfig *rest.Config

var rootCmd = &cobra.Command{
	Use:   "kube-update-image",
	Short: "short kube-update-image",
	Long:  `long kube-update-image`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := kubernetes.NewForConfig(kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		data := "{\"spec\": {\"template\": {\"metadata\": {\"labels\": {\"fake-timestamp\": \"" + strconv.FormatInt(time.Now().Unix(), 10) + "\" }}}}}"
		byteData := *(*[]byte)(unsafe.Pointer(&data))
		deployments, err := clientset.AppsV1().Deployments(namespace).Patch(args[0], types.StrategicMergePatchType, byteData)
		if errors.IsNotFound(err) {
			fmt.Printf("Deployment %s in namespace %s not found\n", args[0], namespace)
			os.Exit(1)
		} else if err != nil {
			panic(err.Error())
		}
		fmt.Printf("deployment/%s updated\n", args[0])
		b, _ := json.Marshal(deployments)
		out := new(bytes.Buffer)
		json.Indent(out, b, "", "    ")
		fmt.Println(out.String())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace")
}

func initConfig() {
	var configString string
	if home, _ := homedir.Dir(); home != "" {
		rootCmd.PersistentFlags().StringVar(&configString, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().StringVar(&configString, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configString)
	if err != nil {
		panic(err.Error())
	}
	kubeconfig = config
}
