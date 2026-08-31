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

// This file implements "payloadless" config loading: loading the job config
// without the pod specs embedded in the job definitions.
//
// Pod specs dominate the memory footprint of a loaded job config: everything
// else about a job is a handful of strings, but the pod spec carries
// containers, args, env, volumes and resource requirements. Components that
// never build a pod (sinker, horologium, tide, ...) still pay for all of it,
// in every replica, for the whole lifetime of the process.
//
// The mechanism has two halves:
//
//  1. Loading: the payloadless loader decodes into mirror types that shadow the
//     `spec` field with podSpecOmitted, a zero-sized type that discards the
//     YAML it is given. The pod spec is never allocated, rather than allocated
//     and dropped afterwards. Everything else (merging, defaulting, regex
//     compilation, validation that does not concern pod specs) runs through the
//     ordinary code path, so a payloadless config is a normally defaulted
//     config that happens to have no pod specs.
//
//  2. Handing out: LoadPayloadless returns a *PayloadlessConfig, a distinct
//     type that exposes the prow config and job metadata but has no way to
//     reach a pod spec. Job views embed the job types through an unexported
//     field and shadow `Spec` with podSpecOmitted, so `job.Spec.Containers`
//     is a compile error rather than a nil pointer dereference. A component
//     asking for a payloadless config cannot silently depend on data that
//     was not loaded.

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"

	"sigs.k8s.io/yaml"
)

// podSpecOmitted takes the place of a *v1.PodSpec that was deliberately not
// loaded. It holds nothing, decodes to nothing, and has no fields or methods,
// so any attempt to use a pod spec that is not there fails to compile.
type podSpecOmitted struct{}

// UnmarshalJSON discards the pod spec instead of allocating it.
func (*podSpecOmitted) UnmarshalJSON([]byte) error { return nil }

// MarshalJSON makes it explicit that no pod spec is being carried, rather than
// serializing an empty object.
func (podSpecOmitted) MarshalJSON() ([]byte, error) { return []byte("null"), nil }

// The mirror types below exist only to be decoded into. Each embeds the real
// job type and shadows its `spec` field: encoding/json resolves the conflict in
// favour of the shallower field, so `spec:` is fed to podSpecOmitted and
// JobBase.Spec is left nil. Embedding rather than duplicating means the mirrors
// cannot drift from the job types they shadow.

type payloadlessPresubmit struct {
	Presubmit
	Spec podSpecOmitted `json:"spec,omitempty"`
}

type payloadlessPostsubmit struct {
	Postsubmit
	Spec podSpecOmitted `json:"spec,omitempty"`
}

type payloadlessPeriodic struct {
	Periodic
	Spec podSpecOmitted `json:"spec,omitempty"`
}

// payloadlessJobConfig mirrors JobConfig. Only the fields that are part of the
// serialized form need to be listed here; the derived fields (AllRepos, the
// ProwYAML getters) are populated by the ordinary load path afterwards.
type payloadlessJobConfig struct {
	Presets           []Preset                           `json:"presets,omitempty"`
	PresubmitsStatic  map[string][]payloadlessPresubmit  `json:"presubmits,omitempty"`
	PostsubmitsStatic map[string][]payloadlessPostsubmit `json:"postsubmits,omitempty"`
	Periodics         []payloadlessPeriodic              `json:"periodics,omitempty"`
	DecorateAllJobs   bool                               `json:"decorate_all_jobs,omitempty"`
	ProwIgnored       *json.RawMessage                   `json:"prow_ignored,omitempty"`
}

// payloadlessConfig mirrors Config for the prow config file, which may carry
// job config of its own.
type payloadlessConfig struct {
	ProwConfig
	payloadlessJobConfig
}

func (p *payloadlessJobConfig) setSourcePath(path string) {
	for repo := range p.PresubmitsStatic {
		for i := range p.PresubmitsStatic[repo] {
			p.PresubmitsStatic[repo][i].SourcePath = path
		}
	}
	for repo := range p.PostsubmitsStatic {
		for i := range p.PostsubmitsStatic[repo] {
			p.PostsubmitsStatic[repo][i].SourcePath = path
		}
	}
	for i := range p.Periodics {
		p.Periodics[i].SourcePath = path
	}
}

// jobConfig unwraps the mirror types into an ordinary JobConfig with nil pod
// specs, so that the rest of the load path (merging, defaulting, validation)
// can operate on it unchanged.
func (p *payloadlessJobConfig) jobConfig() JobConfig {
	jc := JobConfig{
		Presets:         p.Presets,
		DecorateAllJobs: p.DecorateAllJobs,
		ProwIgnored:     p.ProwIgnored,
		PodSpecsOmitted: true,
	}
	if p.PresubmitsStatic != nil {
		jc.PresubmitsStatic = make(map[string][]Presubmit, len(p.PresubmitsStatic))
		for repo, jobs := range p.PresubmitsStatic {
			unwrapped := make([]Presubmit, 0, len(jobs))
			for i := range jobs {
				unwrapped = append(unwrapped, jobs[i].Presubmit)
			}
			jc.PresubmitsStatic[repo] = unwrapped
		}
	}
	if p.PostsubmitsStatic != nil {
		jc.PostsubmitsStatic = make(map[string][]Postsubmit, len(p.PostsubmitsStatic))
		for repo, jobs := range p.PostsubmitsStatic {
			unwrapped := make([]Postsubmit, 0, len(jobs))
			for i := range jobs {
				unwrapped = append(unwrapped, jobs[i].Postsubmit)
			}
			jc.PostsubmitsStatic[repo] = unwrapped
		}
	}
	if p.Periodics != nil {
		jc.Periodics = make([]Periodic, 0, len(p.Periodics))
		for i := range p.Periodics {
			jc.Periodics = append(jc.Periodics, p.Periodics[i].Periodic)
		}
	}
	return jc
}

// The view types below are what a payloadless consumer gets to hold. They embed
// the job type through an unexported field, so every exported field and method
// of the job is promoted and available, while the embedded job itself cannot be
// named from outside this package. `Spec` is shadowed by podSpecOmitted, which
// turns "I forgot this component does not load pod specs" from a nil pointer
// dereference in production into a compile error.

type (
	presubmitView  = Presubmit
	postsubmitView = Postsubmit
	periodicView   = Periodic
)

// PayloadlessPresubmit is a Presubmit whose pod spec was not loaded.
type PayloadlessPresubmit struct {
	*presubmitView
	Spec podSpecOmitted
}

// PayloadlessPostsubmit is a Postsubmit whose pod spec was not loaded.
type PayloadlessPostsubmit struct {
	*postsubmitView
	Spec podSpecOmitted
}

// PayloadlessPeriodic is a Periodic whose pod spec was not loaded.
type PayloadlessPeriodic struct {
	*periodicView
	Spec podSpecOmitted
}

// PayloadlessConfig is a Config that was loaded without job pod specs. It
// deliberately does not embed Config: the prow config is exposed in full, while
// the job config is only reachable through views that cannot yield a pod spec.
type PayloadlessConfig struct {
	ProwConfig

	// PresubmitsStatic contains the presubmits in Prow's main config, without
	// their pod specs. See JobConfig.PresubmitsStatic for the caveats about
	// inrepoconfig.
	PresubmitsStatic map[string][]PayloadlessPresubmit
	// PostsubmitsStatic contains the postsubmits in Prow's main config, without
	// their pod specs. See JobConfig.PostsubmitsStatic for the caveats about
	// inrepoconfig.
	PostsubmitsStatic map[string][]PayloadlessPostsubmit
	// Periodics contains the periodics, without their pod specs.
	Periodics []PayloadlessPeriodic
	// AllRepos contains all repos that have one or more jobs configured or for
	// which a tide query is configured.
	AllRepos map[string]struct{}
}

// AllPeriodics returns the periodics, without their pod specs.
func (c *PayloadlessConfig) AllPeriodics() []PayloadlessPeriodic {
	return c.Periodics
}

// GetPresubmitsStatic returns the presubmits for the given repo, without their
// pod specs.
func (c *PayloadlessConfig) GetPresubmitsStatic(identifier string) []PayloadlessPresubmit {
	return c.PresubmitsStatic[identifier]
}

// GetPostsubmitsStatic returns the postsubmits for the given repo, without their
// pod specs.
func (c *PayloadlessConfig) GetPostsubmitsStatic(identifier string) []PayloadlessPostsubmit {
	return c.PostsubmitsStatic[identifier]
}

// newPayloadlessConfig builds the payloadless view over an already loaded
// config. The views hold pointers into c, so nothing is copied per job.
func newPayloadlessConfig(c *Config) (*PayloadlessConfig, error) {
	if !c.PodSpecsOmitted {
		// Guards against handing out a payloadless view over a config that does
		// carry pod specs: the view would be misleading rather than unsafe, and
		// the memory would not have been saved at all.
		return nil, fmt.Errorf("config was loaded with pod specs, refusing to present it as payloadless")
	}
	p := &PayloadlessConfig{
		ProwConfig: c.ProwConfig,
		AllRepos:   make(map[string]struct{}, len(c.AllRepos)),
	}
	for repo := range c.AllRepos {
		p.AllRepos[repo] = struct{}{}
	}
	if c.PresubmitsStatic != nil {
		p.PresubmitsStatic = make(map[string][]PayloadlessPresubmit, len(c.PresubmitsStatic))
		for repo, jobs := range c.PresubmitsStatic {
			views := make([]PayloadlessPresubmit, 0, len(jobs))
			for i := range jobs {
				views = append(views, PayloadlessPresubmit{presubmitView: &jobs[i]})
			}
			p.PresubmitsStatic[repo] = views
		}
	}
	if c.PostsubmitsStatic != nil {
		p.PostsubmitsStatic = make(map[string][]PayloadlessPostsubmit, len(c.PostsubmitsStatic))
		for repo, jobs := range c.PostsubmitsStatic {
			views := make([]PayloadlessPostsubmit, 0, len(jobs))
			for i := range jobs {
				views = append(views, PayloadlessPostsubmit{postsubmitView: &jobs[i]})
			}
			p.PostsubmitsStatic[repo] = views
		}
	}
	if c.Periodics != nil {
		p.Periodics = make([]PayloadlessPeriodic, 0, len(c.Periodics))
		for i := range c.Periodics {
			p.Periodics = append(p.Periodics, PayloadlessPeriodic{periodicView: &c.Periodics[i]})
		}
	}
	return p, nil
}

// PayloadlessFromConfig strips the pod specs from an already loaded config and
// returns a payloadless view over it.
//
// This is not the memory-saving path: the pod specs were allocated while
// loading, and only the retained footprint shrinks. Use LoadPayloadless to
// avoid reading them in the first place. This exists for tests, and for
// components that hold a full Config and want to hand a payloadless one to a
// part of themselves that has no business looking at pod specs.
//
// The passed config is modified in place: its pod specs are dropped.
func PayloadlessFromConfig(c *Config) *PayloadlessConfig {
	for repo := range c.PresubmitsStatic {
		for i := range c.PresubmitsStatic[repo] {
			c.PresubmitsStatic[repo][i].Spec = nil
		}
	}
	for repo := range c.PostsubmitsStatic {
		for i := range c.PostsubmitsStatic[repo] {
			c.PostsubmitsStatic[repo][i].Spec = nil
		}
	}
	for i := range c.Periodics {
		c.Periodics[i].Spec = nil
	}
	c.PodSpecsOmitted = true
	// newPayloadlessConfig only fails on a config that carries pod specs, which
	// we just made sure is not the case.
	p, _ := newPayloadlessConfig(c)
	return p
}

// LoadPayloadless loads and parses the config at path, without loading the pod
// specs of the jobs. Use it in components that never need to look at a pod
// spec: the pod specs are not read into memory at all, and the returned type
// makes it impossible to depend on them by accident.
func LoadPayloadless(prowConfig, jobConfig string, supplementalProwConfigDirs []string, supplementalProwConfigsFileNameSuffix string, additionals ...func(*PayloadlessConfig) error) (p *PayloadlessConfig, err error) {
	// we never want config loading to take down the prow components.
	defer func() {
		if r := recover(); r != nil {
			p, err = nil, fmt.Errorf("panic loading config: %v\n%s", r, string(debug.Stack()))
		}
	}()
	c, err := loadConfig(prowConfig, jobConfig, supplementalProwConfigDirs, supplementalProwConfigsFileNameSuffix, true, []yaml.JSONOpt{}...)
	if err != nil {
		return nil, err
	}
	if err := c.finalizeJobConfig(); err != nil {
		return nil, err
	}
	if err := c.validateComponentConfig(); err != nil {
		return nil, err
	}
	if err := c.ValidateJobConfig(); err != nil {
		return nil, err
	}
	p, err = newPayloadlessConfig(c)
	if err != nil {
		return nil, err
	}
	for _, additional := range additionals {
		if err := additional(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// PayloadlessGetter returns the current PayloadlessConfig in a thread-safe
// manner. It is the payloadless counterpart of Getter: a component written
// against it cannot be handed a config that carries pod specs, and cannot
// reach a pod spec if it is.
type PayloadlessGetter func() *PayloadlessConfig

// PayloadlessAgent watches the config files and automatically reloads the
// config stored therein, without the job pod specs.
//
// Unlike Agent it does not support subscriptions yet; components that need to
// react to config deltas still use Agent.
type PayloadlessAgent struct {
	mut sync.RWMutex // do not export Lock, etc methods
	c   *PayloadlessConfig
}

// Config returns the latest config. Do not modify the config.
func (ca *PayloadlessAgent) Config() *PayloadlessConfig {
	ca.mut.RLock()
	defer ca.mut.RUnlock()
	return ca.c
}

// Set sets the config. Useful for testing.
func (ca *PayloadlessAgent) Set(c *PayloadlessConfig) {
	ca.mut.Lock()
	defer ca.mut.Unlock()
	ca.c = c
}

// Start will begin polling the config files at the given paths. If the first
// load fails, Start will return the error and abort. Future load failures will
// log the failure message but continue attempting to load.
func (ca *PayloadlessAgent) Start(prowConfig, jobConfig string, additionalProwConfigDirs []string, supplementalProwConfigsFileNameSuffix string, additionals ...func(*PayloadlessConfig) error) error {
	load := func() (*PayloadlessConfig, error) {
		return LoadPayloadless(prowConfig, jobConfig, additionalProwConfigDirs, supplementalProwConfigsFileNameSuffix, additionals...)
	}
	c, err := load()
	if err != nil {
		return err
	}
	ca.Set(c)
	go pollConfig(prowConfig, jobConfig, func() error {
		c, err := load()
		if err != nil {
			return err
		}
		ca.Set(c)
		return nil
	})
	return nil
}
