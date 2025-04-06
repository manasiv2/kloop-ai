package kube

import (
	"context"
	"fmt"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getClient() (*kubernetes.Clientset, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getPods(namespace string) (*corev1.PodList, error) {
	client, err := getClient() // call to k8s
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

/*func CheckPodPhases(namespace string) []PodErrorMetadata {
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
}*/

func GetSuspiciousFromPhase(namespace string) map[string]string {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]string)
	for _, pod := range pods.Items {
		pod_phase := pod.Status.Phase
		if pod_phase != corev1.PodRunning && pod_phase != corev1.PodSucceeded {
			suspiciousPods[string(pod.UID)] = fmt.Sprintf("Pod Phase Issue: %s", pod_phase)
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromContainerStatus(namespace string) map[string]string {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]string)
	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if container.State.Waiting != nil {
				if container.State.Waiting.Reason == "CrashLoopBackOff" || container.State.Waiting.Reason == "ImagePullBackOff" || container.State.Waiting.Reason == "ErrImagePull" {
					suspiciousPods[string(pod.UID)] = fmt.Sprintf("Container Status Issue: %s", container.State.Waiting.Reason)
				}
			} else if container.State.Terminated != nil {
				if container.State.Terminated.ExitCode != 0 {
					suspiciousPods[string(pod.UID)] = fmt.Sprintf("Container Status Issue: %s", container.State.Terminated.Reason)
				}
			} else {
				// init containers -> can check later
			}
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromConditions(namespace string) map[string]string {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]string)
	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse && condition.Reason == "Unschedulable" {
				suspiciousPods[string(pod.UID)] = fmt.Sprintf("Pod Condition Issue: %s, %s", condition.Type, condition.Reason)
			} else if condition.Type == corev1.PodReady || condition.Type == corev1.ContainersReady {
				if condition.Status == corev1.ConditionFalse || condition.Status == corev1.ConditionUnknown {
					suspiciousPods[string(pod.UID)] = fmt.Sprintf("Pod Condition Issue: %s, %s", condition.Type, condition.Reason)
				}
			}
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromEvents(namespace string) map[string]string {
	client, err := getClient()
	if err != nil {
		panic(err.Error())
	}
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]string)
	for _, pod := range pods.Items {
		pod_uid := pod.UID
		events, er := client.CoreV1().Events(namespace).List(
			context.TODO(),
			metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pod_uid),
			},
		)
		if er != nil {
			panic(er)
		}
		for _, event := range events.Items {
			if event.Type == corev1.EventTypeWarning && (event.Reason == "FailedScheduling" || event.Reason == "BackOff" ||
				event.Reason == "FailedMount" || event.Reason == "Unhealthy" || event.Reason == "FailedCreatePodSandBox" || event.Reason == "ErrImagePull" || event.Reason == "ImagePullBackOff") {
				suspiciousPods[string(pod.UID)] = fmt.Sprintf("Event Issue: %s, %s", event.Type, event.Reason)
			}
		}
	}
	return suspiciousPods
}
