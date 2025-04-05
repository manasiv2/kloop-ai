package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodErrorMetadata struct {
	// What went wrong
	Reason  string // e.g., "CrashLoopBackOff", "FailedScheduling", "NotReady", "PendingTooLong"
	Message string // human-readable error message
	Summary string // short, readable description of issue — always populated for LLM

	// Where it happened
	PodName       string
	Namespace     string
	ContainerName string // optional — only relevant for container-level issues

	// Optional diagnostics
	//ExitCode    *int32       // only for terminated containers
	Timestamp *metav1.Time // most recent issue time (can be from status or event)
	//Events      []string     // human-readable event messages, optional -> later
	Labels      map[string]string
	Annotations map[string]string

	// Metadata
	Source string // "ContainerStatus", "PodPhase", "Condition", "EventChecker"
}
