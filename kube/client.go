package kube

import (
	"context"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getPods(namespace string) (*corev1.PodList, error) {
	home, err := os.UserHomeDir() // get home dir path
	if err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config")) // add .kube/config to path
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config) // call to k8s
	if err != nil {
		panic(err.Error())
	}

	var podlist *corev1.PodList

	if namespace == "" {
		podlist, err = client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		podlist, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
	}

	return podlist, nil
}

func CheckPodPhases(namespace string) []PodErrorMetadata {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	var errors []PodErrorMetadata
	for _, pod := range pods.Items {
		pod_phase := pod.Status.Phase
		if pod_phase != "Running" && pod_phase != "Succeeded" {
			poderr := PodErrorMetadata{
				Reason:  string(pod.Status.Phase), // e.g., "CrashLoopBackOff", "FailedScheduling", "NotReady", "PendingTooLong"
				Message: "Pod is not running or succeeded",
				Summary: "Pod stuck in invalid phase",

				// Where it happened
				PodName:       string(pod.Name),
				Namespace:     string(pod.Namespace),
				ContainerName: string(pod.Spec.Containers[0].Name), // optional — only relevant for container-level issues

				// Optional diagnostics
				//ExitCode: 4, //??  // only for terminated containers
				Timestamp: pod.Status.StartTime, // most recent issue time (can be from status or event)
				//Events: //help    // human-readable event messages, optional
				Labels:      pod.Labels,
				Annotations: pod.Annotations,

				// Metadata
				Source: "PodPhase",
			}
			errors = append(errors, poderr)
		}
	}
	return errors
}

//func checkPodConditions(...) []PodErrorMetadata
//func checkPodEvents(...) []PodErrorMetadata
//func checkContainerStatus(namespace string) []PodErrorMetadata
//func sendToLLM([]PodErrorMetadata) string
