package gofalcon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DetectionAPI struct {
	client *Client
}

type QueriesDetectsInput struct {
	Offset *int
	Limit  *int
	Sort   *string
	Filter *string
	Q      *string
}

type QueriesDetectsOutput struct {
	BaseResponse
	Resources []string `json:"resources"`
}

// QueriesDetects retrieves IDs of detection
func (x *DetectionAPI) QueriesDetects(input *QueriesDetectsInput) (*QueriesDetectsOutput, error) {
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
	if input.Filter != nil {
		qs.Add("filter", *input.Filter)
	}
	if input.Q != nil {
		qs.Add("q", *input.Q)
	}

	req := request{
		Method:      "GET",
		Path:        "detects/queries/detects/v1",
		QueryString: qs,
	}

	var output QueriesDetectsOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to query detections")
	}

	Logger.WithFields(logrus.Fields{
		"qs":       qs.Encode(),
		"meta":     output.Meta,
		"returned": len(output.Resources),
	}).Debug("Done QueriesDetects")

	return &output, nil
}

type EntitySummariesInput struct {
	ID []string `json:"ids"`
}

type EntitySummariesOutput struct {
	BaseResponse
	Resources []DetectionResources `json:"resources"`
}

type DetectionBehavior struct {
	AllegedFiletype string `json:"alleged_filetype"`
	BehaviorID      string `json:"behavior_id"`
	Cmdline         string `json:"cmdline"`
	Confidence      int    `json:"confidence"`
	ControlGraphID  string `json:"control_graph_id"`
	DeviceID        string `json:"device_id"`
	Filename        string `json:"filename"`
	IocDescription  string `json:"ioc_description"`
	IocSource       string `json:"ioc_source"`
	IocType         string `json:"ioc_type"`
	IocValue        string `json:"ioc_value"`
	Md5             string `json:"md5"`
	Objective       string `json:"objective"`
	ParentDetails   struct {
		ParentCmdline        string `json:"parent_cmdline"`
		ParentMd5            string `json:"parent_md5"`
		ParentProcessGraphID string `json:"parent_process_graph_id"`
		ParentSha256         string `json:"parent_sha256"`
	} `json:"parent_details"`
	PatternDisposition        int `json:"pattern_disposition"`
	PatternDispositionDetails struct {
		Detect            bool `json:"detect"`
		InddetMask        bool `json:"inddet_mask"`
		Indicator         bool `json:"indicator"`
		KillParent        bool `json:"kill_parent"`
		KillProcess       bool `json:"kill_process"`
		KillSubprocess    bool `json:"kill_subprocess"`
		OperationBlocked  bool `json:"operation_blocked"`
		PolicyDisabled    bool `json:"policy_disabled"`
		ProcessBlocked    bool `json:"process_blocked"`
		QuarantineFile    bool `json:"quarantine_file"`
		QuarantineMachine bool `json:"quarantine_machine"`
		Rooting           bool `json:"rooting"`
		SensorOnly        bool `json:"sensor_only"`
	} `json:"pattern_disposition_details"`
	RuleInstanceID           string    `json:"rule_instance_id"`
	RuleInstanceVersion      int       `json:"rule_instance_version"`
	Scenario                 string    `json:"scenario"`
	Severity                 int       `json:"severity"`
	Sha256                   string    `json:"sha256"`
	Tactic                   string    `json:"tactic"`
	Technique                string    `json:"technique"`
	TemplateInstanceID       string    `json:"template_instance_id"`
	Timestamp                time.Time `json:"timestamp"`
	TriggeringProcessGraphID string    `json:"triggering_process_graph_id"`
	UserID                   string    `json:"user_id"`
	UserName                 string    `json:"user_name"`
}

type DetectionResources struct {
	AdversaryIds     []int               `json:"adversary_ids"`
	AssignedToName   string              `json:"assigned_to_name"`
	AssignedToUID    string              `json:"assigned_to_uid"`
	Behaviors        []DetectionBehavior `json:"behaviors"`
	Cid              string              `json:"cid"`
	CreatedTimestamp time.Time           `json:"created_timestamp"`
	DetectionID      string              `json:"detection_id"`
	Device           DeviceResource      `json:"device"`
	EmailSent        bool                `json:"email_sent"`
	FirstBehavior    time.Time           `json:"first_behavior"`
	Hostinfo         struct {
		ActiveDirectoryDnDisplay []string `json:"active_directory_dn_display"`
		Domain                   string   `json:"domain"`
	} `json:"hostinfo"`
	LastBehavior           time.Time `json:"last_behavior"`
	MaxConfidence          int       `json:"max_confidence"`
	MaxSeverity            int       `json:"max_severity"`
	MaxSeverityDisplayname string    `json:"max_severity_displayname"`
	QuarantinedFiles       []struct {
		ID     string `json:"id"`
		Paths  string `json:"paths"`
		Sha256 string `json:"sha256"`
		State  string `json:"state"`
	} `json:"quarantined_files"`
	SecondsToResolved int    `json:"seconds_to_resolved"`
	SecondsToTriaged  int    `json:"seconds_to_triaged"`
	ShowInUI          bool   `json:"show_in_ui"`
	Status            string `json:"status"`
}

// EntitySummaries retrieves summaries of detection
func (x *DetectionAPI) EntitySummaries(input *EntitySummariesInput) (*EntitySummariesOutput, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to marshal EntitySummaries input")
	}

	req := request{
		Method:  "POST",
		Path:    "detects/entities/summaries/GET/v1",
		Body:    bytes.NewReader(raw),
		Headers: []httpHeader{{"Content-Type", "application/json"}},
	}

	var output EntitySummariesOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to query detections")
	}

	Logger.WithFields(logrus.Fields{
		"input":    input,
		"meta":     output.Meta,
		"returned": len(output.Resources),
	}).Debug("Done EntitySummaries")

	return &output, nil
}
