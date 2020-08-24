/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildJobSpec defines the desired state of BuildJob
type BuildJobSpec struct {
	// The git branch to run this job against
	Branch string `json:"branch"`
	// The commit on the branch to run this job against
	Commit string `json:"commit"`
	// The pipeline this job is associated with
	Pipeline string `json:"pipeline"`
	// The build ID this job is associated with
	Build string `json:"build"`
	// The ID for this job
	Id string `json:"id"`
	// The repo this job needs to sync
	Repo string `json:"repo"`
	// Environment variables to set for this job
	Env []string `json:"env"`
	// Rules for retrying the job if it fails
	RetryRules RetryRules `json:"retry_rules,omitempty"`
	// Paths to automatically collect artifacts from
	ArtifactPaths []string `json:"automatic_artifact_upload_paths,omitempty"`
	// Specified agent query rules, used to pick an environment to run in.
	AgentQueryRules []string `json:"agent_query_rules"`
	// The command to run for the job
	Command []string `json:"command"`
	// Concurrency rules for the job
	Concurrency string `json:"concurrency,omitempty"`
	// Optional metadata to set on this build
	Metadata string `json:"metadata,omitempty"`
}

// Represents the rules for job retries
type RetryRules struct {
	// If true, this job can be manually retried
	Manual bool `json:"manual,omitempty"`
	// The rules to use for deciding if we should auto-retry after a failure
	Automatic AutoRetryRules `json:"automatic,omitempty"`
}

// Specific rules for automatic retries
type AutoRetryRules struct {
	// The exit status that this rule applies to
	ExitStatus string `json:"exit_status,omitempty"`
	// Maximum number of automatic retries for this exit status
	Limit string `json:"limit,omitempty"`
}

// BuildJobStatus defines the observed state of BuildJob
type BuildJobStatus struct {
	// The current state of the build
	State string `json:"state"`
	// When the build was created
	CreatedAt string `json:"created_at"`
	// When the build was cancelled
	CancelledAt string `json:"cancelled_at,omitempty"`
	// When the build started running
	StartedAt string `json:"started_at,omitempty"`
	// When the build finished
	FinishedAt string `json:"finished_at,omitempty"`
	// The exit status returned by the job
	ExitStatus string `json:"exit_status,omitempty"`
	// If the job has finished in a passing state
	Passed bool `json:"passed"`
	// The artifacts collected by the job
	Artifacts []string `json:"artifacts,omitempty"`
}

// BuildJob is the Schema for the buildjobs API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type BuildJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildJobSpec   `json:"spec,omitempty"`
	Status BuildJobStatus `json:"status,omitempty"`
}

// BuildJobList contains a list of BuildJob
// +kubebuilder:object:root=true
type BuildJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BuildJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BuildJob{}, &BuildJobList{})
}
