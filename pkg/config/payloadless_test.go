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
	"reflect"
	"runtime"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

const payloadlessTestProwConfig = `
prowjob_namespace: prowjobs
pod_namespace: pods
sinker:
  resync_period: 1h
  max_prowjob_age: 48h
plank:
  default_decoration_configs:
    "*":
      utility_images:
        clonerefs: gcr.io/k8s-prow/clonerefs:v20250101-abcdef
        initupload: gcr.io/k8s-prow/initupload:v20250101-abcdef
        entrypoint: gcr.io/k8s-prow/entrypoint:v20250101-abcdef
        sidecar: gcr.io/k8s-prow/sidecar:v20250101-abcdef
      gcs_configuration:
        bucket: prow-artifacts
        path_strategy: explicit
      gcs_credentials_secret: service-account
`

// jobConfigYAML generates a job config with jobCount periodics and presubmits,
// each carrying a pod spec of a realistic size.
func jobConfigYAML(jobCount int) string {
	out := "periodics:\n"
	for i := 0; i < jobCount; i++ {
		out += fmt.Sprintf(`- name: periodic-job-%d
  interval: 1h
  decorate: true
  extra_refs:
  - org: kubernetes
    repo: test-infra
    base_ref: master
  spec:
    containers:
    - image: gcr.io/k8s-prow/some-test-image:v20250101-abcdef
      command: [/bin/bash]
      args: ["-c", "make test && make verify && echo done %d"]
      env:
      - name: FOO
        value: bar-%d
      - name: BAZ
        value: qux-%d
      resources:
        requests:
          cpu: "2"
          memory: 4Gi
        limits:
          cpu: "4"
          memory: 8Gi
      volumeMounts:
      - name: creds
        mountPath: /etc/creds
        readOnly: true
    volumes:
    - name: creds
      secret:
        secretName: service-account
`, i, i, i, i)
	}
	out += "presubmits:\n  kubernetes/test-infra:\n"
	for i := 0; i < jobCount; i++ {
		out += fmt.Sprintf(`  - name: pull-job-%d
    always_run: true
    decorate: true
    spec:
      containers:
      - image: gcr.io/k8s-prow/some-test-image:v20250101-abcdef
        command: [/bin/bash]
        args: ["-c", "make test && make verify && echo done %d"]
        env:
        - name: FOO
          value: bar-%d
        resources:
          requests:
            cpu: "2"
            memory: 4Gi
`, i, i, i)
	}
	return out
}

func writePayloadlessTestConfig(t *testing.T, jobCount int) (prowConfig, jobConfig string) {
	t.Helper()
	dir := t.TempDir()
	prowConfig = filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(prowConfig, []byte(payloadlessTestProwConfig), 0644); err != nil {
		t.Fatalf("failed to write prow config: %v", err)
	}
	jobConfig = filepath.Join(dir, "jobs.yaml")
	if err := os.WriteFile(jobConfig, []byte(jobConfigYAML(jobCount)), 0644); err != nil {
		t.Fatalf("failed to write job config: %v", err)
	}
	return prowConfig, jobConfig
}

// TestLoadPayloadlessKeepsEverythingButPodSpecs checks that the payloadless
// loader produces the same config as the normal one, minus the pod specs.
func TestLoadPayloadlessKeepsEverythingButPodSpecs(t *testing.T) {
	prowConfig, jobConfig := writePayloadlessTestConfig(t, 3)

	full, err := Load(prowConfig, jobConfig, nil, "")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	payloadless, err := LoadPayloadless(prowConfig, jobConfig, nil, "")
	if err != nil {
		t.Fatalf("failed to load payloadless config: %v", err)
	}

	if payloadless.ProwJobNamespace != full.ProwJobNamespace {
		t.Errorf("prowjob_namespace: expected %q, got %q", full.ProwJobNamespace, payloadless.ProwJobNamespace)
	}
	if *payloadless.Sinker.ResyncPeriod != *full.Sinker.ResyncPeriod {
		t.Errorf("sinker.resync_period: expected %v, got %v", full.Sinker.ResyncPeriod, payloadless.Sinker.ResyncPeriod)
	}
	if len(payloadless.Periodics) != len(full.Periodics) {
		t.Fatalf("expected %d periodics, got %d", len(full.Periodics), len(payloadless.Periodics))
	}
	for i, p := range payloadless.Periodics {
		// Metadata is available through the view, promoted from the job.
		if p.Name != full.Periodics[i].Name {
			t.Errorf("periodic %d: expected name %q, got %q", i, full.Periodics[i].Name, p.Name)
		}
		if p.GetInterval() != full.Periodics[i].GetInterval() {
			t.Errorf("periodic %d: expected interval %v, got %v", i, full.Periodics[i].GetInterval(), p.GetInterval())
		}
		// Defaulting ran as usual: the decoration config was filled in.
		if p.DecorationConfig == nil {
			t.Errorf("periodic %d: expected the decoration config to be defaulted", i)
		}
		if p.SourcePath != full.Periodics[i].SourcePath {
			t.Errorf("periodic %d: expected source path %q, got %q", i, full.Periodics[i].SourcePath, p.SourcePath)
		}
		// The pod spec was never loaded. From outside the package there is no
		// way to observe this at all: p.Spec is podSpecOmitted, and the
		// embedded job is unexported.
		if p.periodicView.Spec != nil {
			t.Errorf("periodic %d: expected no pod spec to be loaded, got %v", i, p.periodicView.Spec)
		}
	}
	if len(full.Periodics) > 0 && full.Periodics[0].Spec == nil {
		t.Error("the normal loader should still load pod specs")
	}

	presubmits := payloadless.GetPresubmitsStatic("kubernetes/test-infra")
	if expected := len(full.GetPresubmitsStatic("kubernetes/test-infra")); len(presubmits) != expected {
		t.Fatalf("expected %d presubmits, got %d", expected, len(presubmits))
	}
	for i, ps := range presubmits {
		if !ps.AlwaysRun {
			t.Errorf("presubmit %d: expected always_run to be true", i)
		}
		// Compiled regexes and other derived state are there as well.
		if ps.Context == "" {
			t.Errorf("presubmit %d: expected the context to be defaulted", i)
		}
		if ps.presubmitView.Spec != nil {
			t.Errorf("presubmit %d: expected no pod spec to be loaded", i)
		}
	}
}

// TestPayloadlessLoadingSavesMemory measures the retained heap of a loaded
// config with and without pod specs.
func TestPayloadlessLoadingSavesMemory(t *testing.T) {
	const jobCount = 2000
	prowConfig, jobConfig := writePayloadlessTestConfig(t, jobCount)

	retained := func(load func() (any, error)) uint64 {
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		c, err := load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}
		runtime.GC()
		runtime.ReadMemStats(&after)
		// Keep the config alive across the measurement.
		runtime.KeepAlive(c)
		return after.HeapAlloc - before.HeapAlloc
	}

	fullHeap := retained(func() (any, error) { return Load(prowConfig, jobConfig, nil, "") })
	payloadlessHeap := retained(func() (any, error) { return LoadPayloadless(prowConfig, jobConfig, nil, "") })

	t.Logf("%d jobs: full config retains %d KiB, payloadless config retains %d KiB (%.1f%% of full)",
		2*jobCount, fullHeap/1024, payloadlessHeap/1024, 100*float64(payloadlessHeap)/float64(fullHeap))

	if payloadlessHeap >= fullHeap {
		t.Errorf("expected the payloadless config to retain less memory, got %d bytes vs %d bytes", payloadlessHeap, fullHeap)
	}
}

// TestPayloadlessRejectsConfigWithPodSpecs checks the runtime guard that keeps
// a payload-carrying config from being presented as payloadless.
func TestPayloadlessRejectsConfigWithPodSpecs(t *testing.T) {
	if _, err := newPayloadlessConfig(&Config{}); err == nil {
		t.Error("expected an error when building a payloadless view over a config loaded with pod specs")
	}
}

// TestPayloadlessJobConfigMirrorsJobConfig guards the one place where the
// mirror types can drift from the real ones: the job types themselves are
// embedded, so their fields cannot drift, but a job collection newly added to
// JobConfig would silently keep its pod specs.
func TestPayloadlessJobConfigMirrorsJobConfig(t *testing.T) {
	serializedFields := func(t reflect.Type) sets.Set[string] {
		fields := sets.New[string]()
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("json")
			if name, _, _ := strings.Cut(tag, ","); name != "" && name != "-" {
				fields.Insert(name)
			}
		}
		return fields
	}

	real := serializedFields(reflect.TypeOf(JobConfig{}))
	mirror := serializedFields(reflect.TypeOf(payloadlessJobConfig{}))
	if missing := real.Difference(mirror); missing.Len() > 0 {
		t.Errorf("payloadlessJobConfig does not mirror these JobConfig fields: %v. Add them, dropping any pod spec they carry.", sets.List(missing))
	}
	if extra := mirror.Difference(real); extra.Len() > 0 {
		t.Errorf("payloadlessJobConfig has fields JobConfig does not: %v", sets.List(extra))
	}
}
