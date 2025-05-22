package internal

import (
	"context"
	"errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PodsInfo struct {
	PodName string
	Labels  map[string]string
}

func Kubernetes_pod_list(K8sConfig, K8sContext, K8sNamespace, K8sSelector string) (pi []PodsInfo, err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if K8sConfig != "" {
		loadingRules.ExplicitPath = K8sConfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if K8sContext != "" {
		configOverrides.CurrentContext = K8sContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// K8sSelector
	labelmap := make(map[string]string)
	commaChunks := strings.Split(K8sSelector, ",")
	for i := range len(commaChunks) {
		equalChunks := strings.Split(commaChunks[i], "=")
		if len(equalChunks) == 2 {
			labelmap[equalChunks[0]] = equalChunks[1]
		}
	}

	labelSelector := metav1.LabelSelector{MatchLabels: labelmap}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		FieldSelector: "status.phase=Running",
	}

	pods, err := clientset.CoreV1().Pods(K8sNamespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, errors.New("no matching kubernetes pods in running state")
	}

	var podList []PodsInfo
	for num := range len(pods.Items) {
		item := PodsInfo{PodName: pods.Items[num].Name, Labels: pods.Items[num].Labels}
		podList = append(podList, item)
	}
	return podList, nil
}
