package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	var kubeconfig string
	var namespace string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig")
	flag.StringVar(&namespace, "namespace", "", "namespace to watch (empty=all)")
	flag.Parse()

	cfg, err := buildConfig(kubeconfig)
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	var factory informers.SharedInformerFactory
	if namespace == "" {
		factory = informers.NewSharedInformerFactory(clientset, 0)
	} else {
		factory = informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace(namespace))
	}

	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				fmt.Printf("%s - ADDED: %s/%s (IP:%s)\n", time.Now().Format(time.RFC3339), pod.Namespace, pod.Name, pod.Status.PodIP)
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if pod, ok := newObj.(*corev1.Pod); ok {
				fmt.Printf("%s - UPDATED: %s/%s (phase=%s)\n", time.Now().Format(time.RFC3339), pod.Namespace, pod.Name, pod.Status.Phase)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if pod, ok := obj.(*corev1.Pod); ok {
				fmt.Printf("%s - DELETED: %s/%s\n", time.Now().Format(time.RFC3339), pod.Namespace, pod.Name)
			} else {
				// handle tombstone
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					if pod, ok := tombstone.Obj.(*corev1.Pod); ok {
						fmt.Printf("%s - DELETED (tombstone): %s/%s\n", time.Now().Format(time.RFC3339), pod.Namespace, pod.Name)
					}
				}
			}
		},
	})

	stopCh := make(chan struct{})
	factory.Start(stopCh)

	// Wait for caches to sync
	for t, ok := range factory.WaitForCacheSync(stopCh) {
		if !ok {
			log.Fatalf("timed out waiting for caches to sync: %v", t)
		}
	}

	fmt.Println("pod monitor running; watching pod events. (Ctrl+C to stop)")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	close(stopCh)
	fmt.Println("pod monitor exiting")
}
