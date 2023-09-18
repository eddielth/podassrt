package main

import (
	"context"
	"fmt"
	"github.com/eddielth/podassrt/conf"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var configuration conf.Configuration
	err := conf.LoadFromFile(&configuration)
	if err != nil {
		klog.Fatal(err)
	}

	clientset := getK8sClient(&configuration)

	var rsCache = make(map[string][]string, 10)
	for _, restartPolicy := range configuration.RestartPolicy {
		rsCache[restartPolicy.Name] = restartPolicy.Targets
	}

	klog.Info(rsCache)

	namespace := configuration.Common.Namespace
	podsClient := clientset.CoreV1().Pods(namespace)
	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())

	_, controller := cache.NewInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old interface{}, new interface{}) {
			oldPod := old.(*v1.Pod)
			if rsCache[oldPod.Labels[configuration.Common.LabelKey]] == nil {
				return
			}
			for _, target := range rsCache[oldPod.Labels[configuration.Common.LabelKey]] {
				klog.Info(target, " will be deleted from namespace ", namespace)
				pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=%s", configuration.Common.LabelKey, target),
				})
				if err != nil {
					fmt.Printf("Error listing Pods: %v\n", err)
				}
				for _, pod := range pods.Items {
					podsClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
				}
			}

			return
		},
	})

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	// Wait forever
	select {}
}

func getK8sClient(configuration *conf.Configuration) *kubernetes.Clientset {
	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags("", configuration.Common.KubeConfigFile)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatal(err)
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	return clientset
}
