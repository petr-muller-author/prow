/*
Copyright 2025 The Kubernetes Authors.

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
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const sharedDefaultsProwConfig = `
prowjob_namespace: prowjobs
pod_namespace: pods
plank:
  default_decoration_config_entries:
  - config:
      utility_images:
        clonerefs: gcr.io/k8s-prow/clonerefs:v20250101-abcdef
        initupload: gcr.io/k8s-prow/initupload:v20250101-abcdef
        entrypoint: gcr.io/k8s-prow/entrypoint:v20250101-abcdef
        sidecar: gcr.io/k8s-prow/sidecar:v20250101-abcdef
      gcs_configuration:
        bucket: prow-artifacts
        path_strategy: explicit
      gcs_credentials_secret: service-account
  - cluster: trusted-cluster
    config:
      gcs_configuration:
        bucket: trusted-artifacts
        path_strategy: explicit
prowjob_default_entries:
- cluster: "*"
  repo: "*"
  config:
    tenant_id: default-tenant
`

const sharedDefaultsJobConfig = `
periodics:
- name: shared-a
  interval: 1h
  decorate: true
  spec: {containers: [{image: img, command: [/bin/true]}]}
- name: shared-b
  interval: 1h
  decorate: true
  spec: {containers: [{image: img, command: [/bin/true]}]}
- name: on-trusted-cluster
  interval: 1h
  decorate: true
  cluster: trusted-cluster
  spec: {containers: [{image: img, command: [/bin/true]}]}
- name: own-decoration-config
  interval: 1h
  decorate: true
  decoration_config:
    timeout: 3h
  spec: {containers: [{image: img, command: [/bin/true]}]}
presubmits:
  org/repo:
  - name: pull-shared
    always_run: true
    decorate: true
    spec: {containers: [{image: img, command: [/bin/true]}]}
  other-org/other-repo:
  - name: pull-shared-other-repo
    always_run: true
    decorate: true
    spec: {containers: [{image: img, command: [/bin/true]}]}
`

func writeSharedDefaultsConfig(t *testing.T, prowConfig, jobConfig string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	prowConfigPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(prowConfigPath, []byte(prowConfig), 0644); err != nil {
		t.Fatalf("failed to write prow config: %v", err)
	}
	jobConfigPath := filepath.Join(dir, "jobs.yaml")
	if err := os.WriteFile(jobConfigPath, []byte(jobConfig), 0644); err != nil {
		t.Fatalf("failed to write job config: %v", err)
	}
	return prowConfigPath, jobConfigPath
}

// TestJobDefaultsAreShared checks that jobs which end up with the same
// defaulted DecorationConfig and ProwJobDefault share one value, and that jobs
// which do not, do not.
func TestJobDefaultsAreShared(t *testing.T) {
	prowConfig, jobConfig := writeSharedDefaultsConfig(t, sharedDefaultsProwConfig, sharedDefaultsJobConfig)
	c, err := Load(prowConfig, jobConfig, nil, "")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	periodics := map[string]Periodic{}
	for _, p := range c.Periodics {
		periodics[p.Name] = p
	}
	shared := periodics["shared-a"].DecorationConfig
	if shared == nil {
		t.Fatal("expected shared-a to have a defaulted decoration config")
	}
	if other := periodics["shared-b"].DecorationConfig; other != shared {
		t.Error("expected two periodics with the same defaults to share one decoration config")
	}
	// Jobs in different repos still match the same "*" entry, so they share too.
	pullShared := c.GetPresubmitsStatic("org/repo")[0].DecorationConfig
	pullSharedOther := c.GetPresubmitsStatic("other-org/other-repo")[0].DecorationConfig
	if pullShared != shared || pullSharedOther != shared {
		t.Error("expected presubmits matching the same defaults to share the periodics' decoration config")
	}

	// A job matching a different set of entries gets its own value...
	trusted := periodics["on-trusted-cluster"].DecorationConfig
	if trusted == shared {
		t.Error("expected a job on another cluster to get its own decoration config")
	}
	if trusted.GCSConfiguration.Bucket != "trusted-artifacts" {
		t.Errorf("expected the trusted cluster's bucket, got %q", trusted.GCSConfiguration.Bucket)
	}
	// ... and so does a job that brings its own decoration config.
	own := periodics["own-decoration-config"].DecorationConfig
	if own == shared {
		t.Error("expected a job with its own decoration config not to share one")
	}
	if own.Timeout == nil || own.Timeout.Duration.Hours() != 3 {
		t.Errorf("expected the job's own timeout to survive defaulting, got %v", own.Timeout)
	}
	if own.GCSConfiguration.Bucket != "prow-artifacts" {
		t.Errorf("expected the default bucket to be merged in, got %q", own.GCSConfiguration.Bucket)
	}

	// The same holds for prowjob defaults.
	sharedDefault := periodics["shared-a"].ProwJobDefault
	if sharedDefault == nil || sharedDefault.TenantID != "default-tenant" {
		t.Fatalf("expected shared-a to have a defaulted prowjob default, got %v", sharedDefault)
	}
	if periodics["shared-b"].ProwJobDefault != sharedDefault {
		t.Error("expected two periodics with the same defaults to share one prowjob default")
	}
}

// TestSharedJobDefaultsSaveMemory measures the retained heap of a loaded config
// whose jobs get their decoration config defaulted.
func TestSharedJobDefaultsSaveMemory(t *testing.T) {
	const jobCount = 2000
	jobConfig := ""
	for i := 0; i < jobCount; i++ {
		jobConfig += fmt.Sprintf("- name: periodic-job-%d\n  interval: 1h\n  decorate: true\n  spec: {containers: [{image: img, command: [/bin/true]}]}\n", i)
	}
	prowConfigPath, jobConfigPath := writeSharedDefaultsConfig(t, sharedDefaultsProwConfig, "periodics:\n"+jobConfig)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	c, err := Load(prowConfigPath, jobConfigPath, nil, "")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	retained := after.HeapAlloc - before.HeapAlloc
	runtime.KeepAlive(c)

	t.Logf("%d jobs: config retains %d KiB (%d bytes per job)", jobCount, retained/1024, retained/jobCount)
}
