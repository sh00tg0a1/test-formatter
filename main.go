package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type BackupType string

const (
	BackupTypeDB BackupType = "db"
	BackupTypeVM BackupType = "vm"
)

type ParamFormatter struct {
	Job       JobConfig       `json:"job"`
	Source    SourceConfig    `json:"source"`
	Target    TargetConfig    `json:"target"`
	Policy    PolicyConfig    `json:"policy"`
	Execution ExecutionConfig `json:"execution"`
}

type JobConfig struct {
	JobID      string   `json:"job_id"`
	JobName    string   `json:"job_name"`
	JobType    string   `json:"job_type"`
	Priority   string   `json:"priority"`
	TenantID   string   `json:"tenant_id"`
	OperatorID string   `json:"operator_id"`
	Tags       []string `json:"tags"`
}

type SourceConfig struct {
	Resource ResourceInfo `json:"resource"`
	Auth     SourceAuth   `json:"auth"`
}

type ResourceInfo struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ClusterID    string `json:"cluster_id"`
	Namespace    string `json:"namespace"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	VMUUID       string `json:"vm_uuid"`
	Hypervisor   string `json:"hypervisor"`
}

type SourceAuth struct {
	CredentialRef string `json:"credential_ref"`
	AuthMode      string `json:"auth_mode"`
}

type TargetConfig struct {
	Storage   StorageConfig   `json:"storage"`
	Retention RetentionConfig `json:"retention"`
}

type StorageConfig struct {
	Provider     string `json:"provider"`
	Bucket       string `json:"bucket"`
	Path         string `json:"path"`
	Region       string `json:"region"`
	StorageClass string `json:"storage_class"`
	KMSKeyID     string `json:"kms_key_id"`
}

type RetentionConfig struct {
	Mode            string `json:"mode"`
	KeepLast        int    `json:"keep_last"`
	ExpireAfterDays int    `json:"expire_after_days"`
}

type PolicyConfig struct {
	Schedule    ScheduleConfig    `json:"schedule"`
	Consistency ConsistencyConfig `json:"consistency"`
	Security    SecurityConfig    `json:"security"`
}

type ScheduleConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	CronExpr string `json:"cron_expr"`
	Timezone string `json:"timezone"`
	StartAt  string `json:"start_at"`
}

type ConsistencyConfig struct {
	AppConsistent bool   `json:"app_consistent"`
	QuiesceFS     bool   `json:"quiesce_fs"`
	PreScriptRef  string `json:"pre_script_ref"`
	PostScriptRef string `json:"post_script_ref"`
}

type SecurityConfig struct {
	EncryptInTransit bool   `json:"encrypt_in_transit"`
	EncryptAtRest    bool   `json:"encrypt_at_rest"`
	PasswordProtect  bool   `json:"password_protected"`
	PasswordRef      string `json:"password_ref"`
}

type ExecutionConfig struct {
	Retry        RetryConfig        `json:"retry"`
	Performance  PerformanceConfig  `json:"performance"`
	Notification NotificationConfig `json:"notification"`
}

type RetryConfig struct {
	MaxAttempts    int `json:"max_attempts"`
	BackoffSeconds int `json:"backoff_seconds"`
}

type PerformanceConfig struct {
	BandwidthLimitMbps int    `json:"bandwidth_limit_mbps"`
	Parallelism        int    `json:"parallelism"`
	Dedup              bool   `json:"dedup"`
	Compression        string `json:"compression"`
}

type NotificationConfig struct {
	OnSuccess    bool   `json:"on_success"`
	OnFailure    bool   `json:"on_failure"`
	Channel      string `json:"channel"`
	RecipientRef string `json:"recipient_ref"`
}

func main() {
	http.HandleFunc("/param_formatter", paramFormatterHandler)
	http.HandleFunc("/schema", schemaHandler)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func paramFormatterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ParamFormatter
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if err := validateParamFormatter(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(req)
}

func validateParamFormatter(req ParamFormatter) error {
	if req.Source.Resource.ResourceType != string(BackupTypeDB) && req.Source.Resource.ResourceType != string(BackupTypeVM) {
		return httpError("source.resource.resource_type must be one of: db, vm")
	}

	if req.Source.Resource.ResourceType == string(BackupTypeDB) {
		if req.Source.Resource.Host == "" || req.Source.Resource.Port <= 0 || req.Source.Resource.DatabaseName == "" {
			return httpError("for db type, host/port/database_name are required")
		}
	}

	if req.Source.Resource.ResourceType == string(BackupTypeVM) {
		if req.Source.Resource.ClusterID == "" || req.Source.Resource.VMUUID == "" || req.Source.Resource.Hypervisor == "" {
			return httpError("for vm type, cluster_id/vm_uuid/hypervisor are required")
		}
	}

	return nil
}

func httpError(msg string) error {
	return &requestError{Message: msg}
}

type requestError struct {
	Message string
}

func (e *requestError) Error() string {
	return e.Message
}

func schemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   "ParamFormatterAPI",
		"type":    "object",
		"endpoints": map[string]any{
			"POST /param_formatter": map[string]any{
				"description": "输入完整参数对象并原样返回（echo），输入和输出结构一致",
				"request": map[string]any{
					"$ref": "#/definitions/ParamFormatter",
				},
				"response": map[string]any{
					"$ref": "#/definitions/ParamFormatter",
				},
			},
			"GET /schema": map[string]any{
				"description": "返回接口 schema",
			},
		},
		"definitions": map[string]any{
			"ParamFormatter": map[string]any{
				"type":        "object",
				"description": "顶层对象，包含 job/source/target/policy/execution 五个模块",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(schema)
}
