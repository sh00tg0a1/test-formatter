package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func samplePayload(resourceType string) ParamFormatter {
	payload := ParamFormatter{
		Job: JobConfig{
			JobID:      "job-10001",
			JobName:    "nightly-backup",
			JobType:    "full",
			Priority:   "normal",
			TenantID:   "tenant-a",
			OperatorID: "user-ops-01",
			Tags:       []string{"prod", "critical"},
		},
		Source: SourceConfig{
			Resource: ResourceInfo{
				ResourceType: resourceType,
				ResourceID:   "resource-1",
				ResourceName: "name-1",
				ClusterID:    "cluster-a",
				Namespace:    "prod",
				Host:         "10.0.0.21",
				Port:         3306,
				DatabaseName: "orders",
				VMUUID:       "420e9f44-5f11-4f23-9713-40fbb1f66fb1",
				Hypervisor:   "kvm",
			},
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

	if resourceType == string(BackupTypeVM) {
		payload.Source.Resource.Host = ""
		payload.Source.Resource.Port = 0
		payload.Source.Resource.DatabaseName = ""
	}

	if resourceType == string(BackupTypeDB) {
		payload.Source.Resource.ClusterID = ""
		payload.Source.Resource.VMUUID = ""
		payload.Source.Resource.Hypervisor = ""
	}

	return payload
}

func TestParamFormatterEcho_DB(t *testing.T) {
	input := samplePayload(string(BackupTypeDB))
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/param_formatter", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	paramFormatterHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var got ParamFormatter
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if got.Source.Resource.ResourceType != string(BackupTypeDB) {
		t.Fatalf("expected db resource type, got %s", got.Source.Resource.ResourceType)
	}

	if got.Source.Resource.DatabaseName != input.Source.Resource.DatabaseName {
		t.Fatalf("echo mismatch, expected database_name=%s, got %s", input.Source.Resource.DatabaseName, got.Source.Resource.DatabaseName)
	}
}

func TestParamFormatterEcho_VM(t *testing.T) {
	input := samplePayload(string(BackupTypeVM))
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/param_formatter", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	paramFormatterHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var got ParamFormatter
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if got.Source.Resource.ResourceType != string(BackupTypeVM) {
		t.Fatalf("expected vm resource type, got %s", got.Source.Resource.ResourceType)
	}

	if got.Source.Resource.VMUUID != input.Source.Resource.VMUUID {
		t.Fatalf("echo mismatch, expected vm_uuid=%s, got %s", input.Source.Resource.VMUUID, got.Source.Resource.VMUUID)
	}
}

func TestParamFormatterValidate_InvalidType(t *testing.T) {
	input := samplePayload("filesystem")
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/param_formatter", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	paramFormatterHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSchemaHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/schema", nil)
	rec := httptest.NewRecorder()
	schemaHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", got)
	}
}
