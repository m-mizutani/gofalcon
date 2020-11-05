package gofalcon

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DeviceResource struct {
	AgentLoadFlags                string         `json:"agent_load_flags"`
	AgentLocalTime                string         `json:"agent_local_time"`
	AgentVersion                  string         `json:"agent_version"`
	BiosManufacturer              string         `json:"bios_manufacturer"`
	BiosVersion                   string         `json:"bios_version"`
	Cid                           string         `json:"cid"`
	ConfigIDBase                  string         `json:"config_id_base"`
	ConfigIDBuild                 string         `json:"config_id_build"`
	ConfigIDPlatform              string         `json:"config_id_platform"`
	DeviceID                      string         `json:"device_id"`
	DevicePolicies                DevicePolicy   `json:"device_policies"`
	ExternalIP                    string         `json:"external_ip"`
	FirstLoginTimestamp           string         `json:"first_login_timestamp"`
	FirstLoginUser                string         `json:"first_login_user"`
	FirstSeen                     string         `json:"first_seen"`
	Hostname                      string         `json:"hostname"`
	LastLoginTimestamp            string         `json:"last_login_timestamp"`
	LastLoginUser                 string         `json:"last_login_user"`
	LastSeen                      string         `json:"last_seen"`
	LocalIP                       string         `json:"local_ip"`
	MacAddress                    string         `json:"mac_address"`
	MajorVersion                  string         `json:"major_version"`
	Meta                          DeviceMetaData `json:"meta"`
	MinorVersion                  string         `json:"minor_version"`
	ModifiedTimestamp             string         `json:"modified_timestamp"`
	OsVersion                     string         `json:"os_version"`
	Ou                            []string       `json:"ou"`
	PlatformID                    string         `json:"platform_id"`
	PlatformName                  string         `json:"platform_name"`
	Policies                      []Policy       `json:"policies"`
	ProductTypeDesc               string         `json:"product_type_desc"`
	ProvisionStatus               string         `json:"provision_status"`
	SlowChangingModifiedTimestamp string         `json:"slow_changing_modified_timestamp"`
	Status                        string         `json:"status"`
	SystemManufacturer            string         `json:"system_manufacturer"`
	SystemProductName             string         `json:"system_product_name"`
	Tags                          []string       `json:"tags"`
}

type Policy struct {
	Applied      bool   `json:"applied"`
	AppliedDate  string `json:"applied_date"`
	AssignedDate string `json:"assigned_date"`
	PolicyID     string `json:"policy_id"`
	PolicyType   string `json:"policy_type"`
	SettingsHash string `json:"settings_hash"`
}

type DevicePolicy struct {
	GlobalConfig Policy `json:"global_config"`
	Prevention   Policy `json:"prevention"`
	SensorUpdate Policy `json:"sensor_update"`
}

type DeviceMetaData struct {
	Version string `json:"version"`
}

// --------------------------
// Search
//

type DeviceAPI struct {
	client *Client
}

type QueryDevicesFilter struct {
	Key   string
	Value string
}

func (x QueryDevicesFilter) String() string {
	return x.Key + ":" + x.Value
}

type QueryDevicesInput struct {
	Offset  *int
	Limit   *int
	Sort    *string
	Filters []QueryDevicesFilter
}

type QueryDevicesOutput struct {
	BaseResponse
	Resources []string `json:"resources"`
}

// QueryDevices searches hosts in your environment by platform, hostname, IP, and other criteria.
func (x *DeviceAPI) QueryDevices(input *QueryDevicesInput) (*QueryDevicesOutput, error) {
	qs := url.Values{}
	if input.Offset != nil {
		qs.Add("offset", fmt.Sprintf("%d", *input.Offset))
	}
	if input.Limit != nil {
		qs.Add("limit", fmt.Sprintf("%d", *input.Limit))
	}
	if input.Sort != nil {
		qs.Add("sort", *input.Sort)
	}
	if len(input.Filters) > 0 {
		var filters []string
		for _, filter := range input.Filters {
			filters = append(filters, filter.String())
		}
		qs.Add("filter", strings.Join(filters, "+"))
	}

	req := Request{
		Path:        "devices/queries/devices/v1",
		QueryString: qs,
	}

	var output QueryDevicesOutput
	if err := x.client.SendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to QueryDevice")
	}

	Logger.WithFields(logrus.Fields{
		"qs":       qs.Encode(),
		"meta":     output.Meta,
		"returned": len(output.Resources),
	}).Debug("Done QueryDevices")

	return &output, nil
}

type EntityDevicesInput struct {
	ID []string
}

type EntityDevicesOutput struct {
	BaseResponse
	Resources []DeviceResource `json:"resources"`
}

// EntityDevices gets details on one or more hosts by providing agent IDs (AID)
func (x *DeviceAPI) EntityDevices(input *EntityDevicesInput) (*EntityDevicesOutput, error) {
	qs := url.Values{}
	for _, id := range input.ID {
		qs.Add("ids", id)
	}

	req := Request{
		Path:        "devices/entities/devices/v1",
		QueryString: qs,
	}

	var output EntityDevicesOutput
	if err := x.client.SendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to EntityDevices")
	}

	Logger.WithFields(logrus.Fields{
		"qs":       qs.Encode(),
		"meta":     output.Meta,
		"returned": len(output.Resources),
	}).Debug("Done EntityDevices")

	return &output, nil
}
