package kube

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodErrorMetadata struct {
	// What went wrong
	Summary string // One-liner for LLM (e.g. "Pod is in CrashLoopBackOff due to image pull error")
	Reason  string // Precise signal (e.g. "CrashLoopBackOff", "FailedScheduling", etc)
	Message string // Verbose human-readable explanation

	// Where it happened
	PodName       string
	Namespace     string
	ContainerName string // Optional: only for container-level issues

	// Optional diagnostics
	Timestamp      *metav1.Time             // When it happened
	ExitCode       *int32                   // For container termination issues
	EventMessages  []string                 // Key warning events (for event checker)
	Conditions     []corev1.PodCondition    // For deep analysis of pod readiness/scheduling
	Status         corev1.PodPhase          // Pod phase at time of issue
	InitStatus     []corev1.ContainerStatus // Optional: init container states
	ContainerState *corev1.ContainerState   // Optional: specific container state if applicable

	Labels      map[string]string
	Annotations map[string]string

	// Origin
	Source string // "ContainerStatus", "PodPhase", "Condition", "EventChecker"
}
