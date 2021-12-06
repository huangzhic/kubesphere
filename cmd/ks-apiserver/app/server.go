/*
Copyright 2019 The KubeSphere Authors.

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

package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"kubesphere.io/kubesphere/cmd/ks-apiserver/app/options"
	apiserverconfig "kubesphere.io/kubesphere/pkg/apiserver/config"
	"kubesphere.io/kubesphere/pkg/utils/term"
	"kubesphere.io/kubesphere/pkg/version"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	// 从文件中读取配置
	conf, err := apiserverconfig.TryLoadFromDisk()
	if err == nil {
		s = &options.ServerRunOptions{
			GenericServerRunOptions: s.GenericServerRunOptions,
			Config:                  conf,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	//创建一个command
	cmd := &cobra.Command{
		//使用信息
		Use: "ks-apiserver",
		//长描述
		Long: `The KubeSphere API server validates and configures data for the API objects. 
The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				//聚合错误
				return utilerrors.NewAggregate(errs)
			}

			return Run(s, signals.SetupSignalHandler())
		},
		//当发生错误时静默使用
		SilenceUsage: true,
	}

	// 得到cmd的*Flagset
	fs := cmd.Flags()
	// 得到NamedFlagSets
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		//全部加到fs中
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	// 返回用户终端的宽度和高度
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	// 自定义帮助选项
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	//创建version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere ks-apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func Run(s *options.ServerRunOptions, ctx context.Context) error {
	//得到api server
	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(ctx.Done())
	if err != nil {
		return err
	}

	return apiserver.Run(ctx)
}
