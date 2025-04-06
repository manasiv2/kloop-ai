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

func GetSuspiciousFromPhase(namespace string) map[string]PodErrorMetadata {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]PodErrorMetadata)
	var addpod bool = false
	for _, pod := range pods.Items {
		pod_phase := pod.Status.Phase
		if pod_phase != corev1.PodRunning && pod_phase != corev1.PodSucceeded {
			//suspiciousPods[string(pod.UID)] = fmt.Sprintf("Pod Phase Issue: %s", pod_phase)
			addpod = true
		}
		if addpod {
			addpod = false
			poderr := PodErrorMetadata{
				Summary:     "Pod Phase Issue",
				Reason:      string(pod.Status.Phase),
				Message:     "Pod is not running or succeeded",
				PodName:     pod.Name,
				Namespace:   pod.Namespace,
				Status:      pod.Status.Phase,
				Timestamp:   pod.Status.StartTime,
				Labels:      pod.Labels,
				Annotations: pod.Annotations,
				Conditions:  pod.Status.Conditions,
				Source:      "PodPhase",
			}

			suspiciousPods[string(pod.UID)] = poderr
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromContainerStatus(namespace string) map[string]PodErrorMetadata {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]PodErrorMetadata)

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			var reason, msg string
			var containerState *corev1.ContainerState = &container.State

			if container.State.Waiting != nil {
				reason = container.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					msg = container.State.Waiting.Message
				}
			} else if container.State.Terminated != nil && container.State.Terminated.ExitCode != 0 {
				reason = container.State.Terminated.Reason
				msg = container.State.Terminated.Message
			}

			if reason != "" {
				poderr := PodErrorMetadata{
					Summary:        "Container Status Issue",
					Reason:         reason,
					Message:        msg,
					PodName:        pod.Name,
					Namespace:      pod.Namespace,
					ContainerName:  container.Name,
					Timestamp:      pod.Status.StartTime,
					Labels:         pod.Labels,
					Annotations:    pod.Annotations,
					Status:         pod.Status.Phase,
					ContainerState: containerState,
					Conditions:     pod.Status.Conditions,
					Source:         "ContainerStatus",
				}
				suspiciousPods[string(pod.UID)] = poderr
			}
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromConditions(namespace string) map[string]PodErrorMetadata {
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]PodErrorMetadata)

	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if (condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse && condition.Reason == "Unschedulable") ||
				((condition.Type == corev1.PodReady || condition.Type == corev1.ContainersReady) && (condition.Status == corev1.ConditionFalse || condition.Status == corev1.ConditionUnknown)) {

				poderr := PodErrorMetadata{
					Summary:     "Pod Condition Issue",
					Reason:      condition.Reason,
					Message:     condition.Message,
					PodName:     pod.Name,
					Namespace:   pod.Namespace,
					Timestamp:   &condition.LastTransitionTime,
					Labels:      pod.Labels,
					Annotations: pod.Annotations,
					Status:      pod.Status.Phase,
					Conditions:  pod.Status.Conditions,
					Source:      "PodCondition",
				}
				suspiciousPods[string(pod.UID)] = poderr
				break
			}
		}
	}
	return suspiciousPods
}

func GetSuspiciousFromEvents(namespace string) map[string]PodErrorMetadata {
	client, err := getClient()
	if err != nil {
		panic(err)
	}
	pods, err := getPods(namespace)
	if err != nil {
		panic(err)
	}
	suspiciousPods := make(map[string]PodErrorMetadata)

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
		var warnings []string
		for _, event := range events.Items {
			if event.Type == corev1.EventTypeWarning && (event.Reason == "FailedScheduling" || event.Reason == "BackOff" ||
				event.Reason == "FailedMount" || event.Reason == "Unhealthy" || event.Reason == "FailedCreatePodSandBox" || event.Reason == "ErrImagePull" || event.Reason == "ImagePullBackOff") {

				warnings = append(warnings, fmt.Sprintf("%s: %s", event.Reason, event.Message))
			}
		}

		if len(warnings) > 0 {
			poderr := PodErrorMetadata{
				Summary:       "Warning Events Found",
				Reason:        "EventType: Warning",
				Message:       "Pod has one or more warning-level events",
				PodName:       pod.Name,
				Namespace:     pod.Namespace,
				Timestamp:     pod.Status.StartTime,
				Labels:        pod.Labels,
				Annotations:   pod.Annotations,
				Status:        pod.Status.Phase,
				EventMessages: warnings,
				Conditions:    pod.Status.Conditions,
				Source:        "EventChecker",
			}
			suspiciousPods[string(pod.UID)] = poderr
		}
	}
	return suspiciousPods
}
