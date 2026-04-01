package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type BackupType string

const (
	BackupTypeDB BackupType = "db"
	BackupTypeVM BackupType = "vm"
)

type ParamFormatterRequest struct {
	BackupType BackupType `json:"backup_type"`
}

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
	ClusterID    string `json:"cluster_id,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	DatabaseName string `json:"database_name,omitempty"`
	VMUUID       string `json:"vm_uuid,omitempty"`
	Hypervisor   string `json:"hypervisor,omitempty"`
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

	var req ParamFormatterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	backupType := BackupType(strings.ToLower(string(req.BackupType)))
	if backupType != BackupTypeDB && backupType != BackupTypeVM {
		http.Error(w, "backup_type must be one of: db, vm", http.StatusBadRequest)
		return
	}

	resp := buildTemplate(backupType)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func buildTemplate(t BackupType) ParamFormatter {
	base := ParamFormatter{
		Job: JobConfig{
			JobID:      "job-10001",
			JobName:    "nightly-backup",
			JobType:    "full",
			Priority:   "normal",
			TenantID:   "tenant-a",
			OperatorID: "user-ops-01",
			Tags:       []string{"prod", "critical", string(t)},
		},
		Source: SourceConfig{
			Auth: SourceAuth{
				CredentialRef: "credential/default",
				AuthMode:      "token",
			},
		},
		Target: TargetConfig{
			Storage: StorageConfig{
				Provider:     "s3",
				Bucket:       "backup-bucket",
				Path:         "/daily",
				Region:       "us-east-1",
				StorageClass: "standard",
				KMSKeyID:     "kms-key-001",
			},
			Retention: RetentionConfig{
				Mode:            "days",
				KeepLast:        30,
				ExpireAfterDays: 180,
			},
		},
		Policy: PolicyConfig{
			Schedule: ScheduleConfig{
				Enabled:  true,
				Type:     "cron",
				CronExpr: "0 2 * * *",
				Timezone: "UTC",
				StartAt:  "2026-04-01T02:00:00Z",
			},
			Consistency: ConsistencyConfig{
				AppConsistent: true,
				QuiesceFS:     true,
				PreScriptRef:  "script/pre-freeze",
				PostScriptRef: "script/post-thaw",
			},
			Security: SecurityConfig{
				EncryptInTransit: true,
				EncryptAtRest:    true,
				PasswordProtect:  false,
				PasswordRef:      "",
			},
		},
		Execution: ExecutionConfig{
			Retry: RetryConfig{
				MaxAttempts:    3,
				BackoffSeconds: 15,
			},
			Performance: PerformanceConfig{
				BandwidthLimitMbps: 500,
				Parallelism:        8,
				Dedup:              true,
				Compression:        "lz4",
			},
			Notification: NotificationConfig{
				OnSuccess:    true,
				OnFailure:    true,
				Channel:      "webhook",
				RecipientRef: "notify/ops-webhook",
			},
		},
	}

	if t == BackupTypeDB {
		base.Source.Resource = ResourceInfo{
			ResourceType: "db",
			ResourceID:   "db-001",
			ResourceName: "orders-mysql",
			Host:         "10.0.0.21",
			Port:         3306,
			DatabaseName: "orders",
		}
	} else {
		base.Source.Resource = ResourceInfo{
			ResourceType: "vm",
			ResourceID:   "vm-001",
			ResourceName: "billing-vm-01",
			ClusterID:    "cluster-a",
			Namespace:    "prod",
			VMUUID:       "420e9f44-5f11-4f23-9713-40fbb1f66fb1",
			Hypervisor:   "kvm",
		}
	}

	return base
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
				"description": "根据 backup_type 返回对应的三层嵌套备份参数模板",
				"request": map[string]any{
					"type":     "object",
					"required": []string{"backup_type"},
					"properties": map[string]any{
						"backup_type": map[string]any{
							"type": "string",
							"enum": []string{"db", "vm"},
						},
					},
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
