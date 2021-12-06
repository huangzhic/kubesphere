/*
Copyright 2020 The KubeSphere Authors.

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

package config

import (
	"fmt"
	"reflect"
	"strings"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization"

	"github.com/spf13/viper"

	networkv1alpha1 "kubesphere.io/api/network/v1alpha1"

	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
	"kubesphere.io/kubesphere/pkg/simple/client/gpu"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/kubeedge"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring/prometheus"
	"kubesphere.io/kubesphere/pkg/simple/client/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/network"
	"kubesphere.io/kubesphere/pkg/simple/client/notification"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
)

// Package config保存运行KubeSphere组件的配置
//
// Config 可以来自于命令行flags和配置文件
// 命令行flags优先级高于配置文件
// 但如果组件的Endpoint/Host/APIServer 为空,
// 那个组件的所有命令行flags将被忽视，转而使用配置文件。
// 例如，我们有一个配置文件
//
// mysql:
//   host: mysql.kubesphere-system.svc
//   username: root
//   password: password
//
// 同时，具有如下命令行标志：
//
// --mysql-host mysql.openpitrix-system.svc --mysql-username king --mysql-password 1234
//
//我们将使用 `king:1234@mysql.openpitrix-system.svc` 来自命令行标志
// 但如果命令行flags如下:
//
// --mysql-username root --mysql-password password
//
// 我们将使用`root:password@mysql.kubesphere-system.svc` as input,
// 因为命令行flags中缺少mysql-host，该命令行的所有其他flags将被忽略。

const (
	// 默认配置名
	defaultConfigurationName = "kubesphere"

	// 默认配置路径
	defaultConfigurationPath = "/etc/kubesphere"
)

// Config 定义apiserver处理外部服务所需的一切
type Config struct {
	DevopsOptions         *jenkins.Options        `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	SonarQubeOptions      *sonarqube.Options      `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	KubernetesOptions     *k8s.KubernetesOptions  `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	ServiceMeshOptions    *servicemesh.Options    `json:"servicemesh,omitempty" yaml:"servicemesh,omitempty" mapstructure:"servicemesh"`
	NetworkOptions        *network.Options        `json:"network,omitempty" yaml:"network,omitempty" mapstructure:"network"`
	LdapOptions           *ldap.Options           `json:"-,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	RedisOptions          *cache.Options          `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	S3Options             *s3.Options             `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	OpenPitrixOptions     *openpitrix.Options     `json:"openpitrix,omitempty" yaml:"openpitrix,omitempty" mapstructure:"openpitrix"`
	MonitoringOptions     *prometheus.Options     `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
	LoggingOptions        *logging.Options        `json:"logging,omitempty" yaml:"logging,omitempty" mapstructure:"logging"`
	AuthenticationOptions *authentication.Options `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	AuthorizationOptions  *authorization.Options  `json:"authorization,omitempty" yaml:"authorization,omitempty" mapstructure:"authorization"`
	MultiClusterOptions   *multicluster.Options   `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
	EventsOptions         *events.Options         `json:"events,omitempty" yaml:"events,omitempty" mapstructure:"events"`
	AuditingOptions       *auditing.Options       `json:"auditing,omitempty" yaml:"auditing,omitempty" mapstructure:"auditing"`
	AlertingOptions       *alerting.Options       `json:"alerting,omitempty" yaml:"alerting,omitempty" mapstructure:"alerting"`
	NotificationOptions   *notification.Options   `json:"notification,omitempty" yaml:"notification,omitempty" mapstructure:"notification"`
	KubeEdgeOptions       *kubeedge.Options       `json:"kubeedge,omitempty" yaml:"kubeedge,omitempty" mapstructure:"kubeedge"`
	MeteringOptions       *metering.Options       `json:"metering,omitempty" yaml:"metering,omitempty" mapstructure:"metering"`
	GatewayOptions        *gateway.Options        `json:"gateway,omitempty" yaml:"gateway,omitempty" mapstructure:"gateway"`
	GPUOptions            *gpu.Options            `json:"gpu,omitempty" yaml:"gpu,omitempty" mapstructure:"gpu"`
}

// 创建一个默认非空Config
func New() *Config {
	return &Config{
		DevopsOptions:         jenkins.NewDevopsOptions(),
		SonarQubeOptions:      sonarqube.NewSonarQubeOptions(),
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		ServiceMeshOptions:    servicemesh.NewServiceMeshOptions(),
		NetworkOptions:        network.NewNetworkOptions(),
		LdapOptions:           ldap.NewOptions(),
		RedisOptions:          cache.NewRedisOptions(),
		S3Options:             s3.NewS3Options(),
		OpenPitrixOptions:     openpitrix.NewOptions(),
		MonitoringOptions:     prometheus.NewPrometheusOptions(),
		AlertingOptions:       alerting.NewAlertingOptions(),
		NotificationOptions:   notification.NewNotificationOptions(),
		LoggingOptions:        logging.NewLoggingOptions(),
		AuthenticationOptions: authentication.NewOptions(),
		AuthorizationOptions:  authorization.NewOptions(),
		MultiClusterOptions:   multicluster.NewOptions(),
		EventsOptions:         events.NewEventsOptions(),
		AuditingOptions:       auditing.NewAuditingOptions(),
		KubeEdgeOptions:       kubeedge.NewKubeEdgeOptions(),
		MeteringOptions:       metering.NewMeteringOptions(),
		GatewayOptions:        gateway.NewGatewayOptions(),
		GPUOptions:            gpu.NewGPUOptions(),
	}
}

// TryLoadFromDisk 服务器启动后从默认路径加载配置
// return nil error 如果配置文件不存在
func TryLoadFromDisk() (*Config, error) {
	// 配置文件的名字
	viper.SetConfigName(defaultConfigurationName)
	// 配置文件路径
	viper.AddConfigPath(defaultConfigurationPath)

	// 添加多个配置文件路径
	// 从当前工作目录加载配置文件，仅用于debug时使用
	viper.AddConfigPath(".")

	// 从环境变量中读取
	viper.SetEnvPrefix("kubesphere")
	// 开启读取环境变量
	viper.AutomaticEnv()
	// 将环境变量中的.替换成_
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 如果配置文件没找到
			return nil, err
		} else {
			// 如果配置文件找到了但发生了其他错误
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := New()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	conf.stripEmptyOptions()
	result := make(map[string]bool, 0)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if name == "network" {
			ippoolName := "network.ippool"
			nsnpName := "network"
			networkTopologyName := "network.topology"
			if conf.NetworkOptions == nil {
				result[nsnpName] = false
				result[ippoolName] = false
			} else {
				if conf.NetworkOptions.EnableNetworkPolicy {
					result[nsnpName] = true
				} else {
					result[nsnpName] = false
				}

				if conf.NetworkOptions.IPPoolType == networkv1alpha1.IPPoolTypeNone {
					result[ippoolName] = false
				} else {
					result[ippoolName] = true
				}

				if conf.NetworkOptions.WeaveScopeHost == "" {
					result[networkTopologyName] = false
				} else {
					result[networkTopologyName] = true
				}
			}
			continue
		}

		if name == "openpitrix" {
			// openpitrix is always true
			result[name] = true
			if conf.OpenPitrixOptions == nil {
				result["openpitrix.appstore"] = false
			} else {
				result["openpitrix.appstore"] = !conf.OpenPitrixOptions.AppStoreConfIsEmpty()
			}
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}

// Remove invalid options before serializing to json or yaml
func (conf *Config) stripEmptyOptions() {

	if conf.RedisOptions != nil && conf.RedisOptions.Host == "" {
		conf.RedisOptions = nil
	}

	if conf.DevopsOptions != nil && conf.DevopsOptions.Host == "" {
		conf.DevopsOptions = nil
	}

	if conf.MonitoringOptions != nil && conf.MonitoringOptions.Endpoint == "" {
		conf.MonitoringOptions = nil
	}

	if conf.SonarQubeOptions != nil && conf.SonarQubeOptions.Host == "" {
		conf.SonarQubeOptions = nil
	}

	if conf.LdapOptions != nil && conf.LdapOptions.Host == "" {
		conf.LdapOptions = nil
	}

	if conf.NetworkOptions != nil && conf.NetworkOptions.IsEmpty() {
		conf.NetworkOptions = nil
	}

	if conf.ServiceMeshOptions != nil && conf.ServiceMeshOptions.IstioPilotHost == "" &&
		conf.ServiceMeshOptions.ServicemeshPrometheusHost == "" &&
		conf.ServiceMeshOptions.JaegerQueryHost == "" {
		conf.ServiceMeshOptions = nil
	}

	if conf.S3Options != nil && conf.S3Options.Endpoint == "" {
		conf.S3Options = nil
	}

	if conf.AlertingOptions != nil && conf.AlertingOptions.Endpoint == "" &&
		conf.AlertingOptions.PrometheusEndpoint == "" && conf.AlertingOptions.ThanosRulerEndpoint == "" {
		conf.AlertingOptions = nil
	}

	if conf.LoggingOptions != nil && conf.LoggingOptions.Host == "" {
		conf.LoggingOptions = nil
	}

	if conf.NotificationOptions != nil && conf.NotificationOptions.Endpoint == "" {
		conf.NotificationOptions = nil
	}

	if conf.MultiClusterOptions != nil && !conf.MultiClusterOptions.Enable {
		conf.MultiClusterOptions = nil
	}

	if conf.EventsOptions != nil && conf.EventsOptions.Host == "" {
		conf.EventsOptions = nil
	}

	if conf.AuditingOptions != nil && conf.AuditingOptions.Host == "" {
		conf.AuditingOptions = nil
	}

	if conf.KubeEdgeOptions != nil && conf.KubeEdgeOptions.Endpoint == "" {
		conf.KubeEdgeOptions = nil
	}

	if conf.GPUOptions != nil && len(conf.GPUOptions.Kinds) == 0 {
		conf.GPUOptions = nil
	}
}
