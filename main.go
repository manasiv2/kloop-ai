package main

import (
	"fmt"

	"github.com/manasiv2/kloop-ai/kube"
)

func main() {
	namespace := "" // or "default" for a specific namespace

	fmt.Println("🔍 Running Phase Check...")
	phaseErrors := kube.GetSuspiciousFromPhase(namespace)
	fmt.Printf("Found %d phase issues\n", len(phaseErrors))
	for _, err := range phaseErrors {
		fmt.Printf("[Phase] %s: %s\n", err.PodName, err.Summary)
	}

	fmt.Println("\n🔍 Running Container Status Check...")
	containerErrors := kube.GetSuspiciousFromContainerStatus(namespace)
	fmt.Printf("Found %d container status issues\n", len(containerErrors))
	for _, err := range containerErrors {
		fmt.Printf("[Container] %s: %s\n", err.PodName, err.Summary)
	}

	fmt.Println("\n🔍 Running Condition Check...")
	conditionErrors := kube.GetSuspiciousFromConditions(namespace)
	fmt.Printf("Found %d condition issues\n", len(conditionErrors))
	for _, err := range conditionErrors {
		fmt.Printf("[Condition] %s: %s\n", err.PodName, err.Summary)
	}

	fmt.Println("\n🔍 Running Event Check...")
	eventErrors := kube.GetSuspiciousFromEvents(namespace)
	fmt.Printf("Found %d event issues\n", len(eventErrors))
	for _, err := range eventErrors {
		fmt.Printf("[Event] %s: %s\n", err.PodName, err.Summary)
	}
}
