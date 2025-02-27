/*
Copyright © 2024 masteryyh <yyh991013@163.com>

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
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var config Config

type NetworkStack string

const (
	IPv4 NetworkStack = "IPv4"
	IPv6 NetworkStack = "IPv6"
)

type AddressDetectionType string

const (
	AddressDetectionIface      AddressDetectionType = "Interface"
	AddressDetectionThirdParty AddressDetectionType = "ThirdParty"
)

type LocalAddressPolicy string

const (
	LocalAddressPolicyIgnore LocalAddressPolicy = "Ignore"
	LocalAddressPolicyAllow  LocalAddressPolicy = "Allow"
	LocalAddressPolicyPrefer LocalAddressPolicy = "Prefer"
)

type DNSProvider string

const (
	DNSProviderCloudflare  DNSProvider = "Cloudflare"
	DNSProviderAliCloud    DNSProvider = "AliCloud"
	DNSProviderDNSPod      DNSProvider = "DNSPod"
	DNSProviderHuaweiCloud DNSProvider = "HuaweiCloud"
	DNSProviderJDCloud     DNSProvider = "JDCloud"
)

// DNSProviderSpec is the specification of DNS provider, currently only Cloudflare
// is supported
type DNSProviderSpec struct {
	// Name is the name of DNS provider
	Name DNSProvider `json:"name" yaml:"name"`

	Cloudflare *CloudflareSpec `json:"cloudflare,omitempty" yaml:"cloudflare,omitempty"`

	AliCloud *AliCloudSpec `json:"alicloud,omitempty" yaml:"alicloud,omitempty"`

	DNSPod *DNSPodSpec `json:"dnspod,omitempty" yaml:"dnspod,omitempty"`

	Huawei *HuaweiCloudSpec `json:"huawei,omitempty" yaml:"huawei,omitempty"`

	JD *JDCloudSpec `json:"jd,omitempty" yaml:"jd,omitempty"`
}

// NetworkInterfaceDetectionSpec defines how should we get IP address from an interface
// By default the first address detected will be used
type NetworkInterfaceDetectionSpec struct {
	// Name is the name of interface
	Name string `json:"name" yaml:"name"`
}

// ThirdPartyServiceSpec defines how should we access third party API to get our IP address
type ThirdPartyServiceSpec struct {
	// URL is the URL of third-party API
	URL string `json:"url" yaml:"url"`

	// JsonPath is the path to the address if data returned by API is JSON-formatted
	JsonPath *string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`

	// Params will be added to the URL
	Params *map[string]string `json:"params,omitempty" yaml:"params,omitempty"`

	// Headers will be added to the request header if not empty
	Headers *map[string]string `json:"customHeaders,omitempty" yaml:"customHeaders,omitempty"`

	// Username is the username for HTTP basic authentication if required
	Username *string `json:"username,omitempty" yaml:"username,omitempty"`

	// Password is the password for HTTP basic authentication if required
	Password *string `json:"password,omitempty" yaml:"password,omitempty"`
}

// AddressDetectionSpec defines how should we detect current IP address
type AddressDetectionSpec struct {
	// Type is the type of address detection method
	// Currently we can acquire address by network interface or 3rd-party API
	Type AddressDetectionType `json:"type" yaml:"type"`

	// LocalAddressPolicy defines how should we process addresses
	// LocalAddressPolicyIgnore means the operation would fail when no public address presents on the interface
	// LocalAddressPolicyAllow means local addresses will be used for DNS record, but only if no public address presents on the interface
	// LocalAddressPolicyPrefer means local addresses will be used for DNS record even public address presents on the interface
	LocalAddressPolicy *LocalAddressPolicy `json:"localAddressPolicy,omitempty" yaml:"localAddressPolicy,omitempty"`

	Interface *NetworkInterfaceDetectionSpec `json:"interface,omitempty" yaml:"interface,omitempty"`

	API *ThirdPartyServiceSpec `json:"api,omitempty" yaml:"api,omitempty"`
}

// DDNSSpec is the specification of DDNS service
type DDNSSpec struct {
	// Name is the name of the specification
	Name string `json:"name" yaml:"name"`

	// Domain is the domain of user
	Domain string `json:"domain" yaml:"domain"`

	// Subdomain is the subdomain to update, use "@" if no subdomain is used
	Subdomain string `json:"subdomain" yaml:"subdomain"`

	// Stack determines if IPv4 or IPv6 is used
	Stack NetworkStack `json:"stack" yaml:"stack"`

	// Cron is the cron expression about how should we schedule this task
	Cron string `json:"cron" yaml:"cron"`

	Detection AddressDetectionSpec `json:"detection" yaml:"detection"`

	Provider DNSProviderSpec `json:"provider" yaml:"provider"`
}

// Config is the configuration of this application
type Config struct {
	DDNS []*DDNSSpec `json:"ddns,omitempty" yaml:"ddns,omitempty"`
}

func ReadConfigOrGet(path string) (*Config, error) {
	if len(config.DDNS) > 0 {
		return &config, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("config path points to a directory")
	}

	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("config path points to an unknown file type")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fileType := parts[len(parts)-1]
	switch fileType {
	case "yaml", "yml":
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	case "json":
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("config path points to an unknown file type")
	}

	return &config, nil
}
