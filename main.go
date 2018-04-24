package main

import (
	"fmt"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	discocache "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/sample-controller/pkg/signals"

	clientset "github.com/lilic/kube-start-stop/pkg/client/clientset/versioned"
	informers "github.com/lilic/kube-start-stop/pkg/client/informers/externalversions"
	"github.com/lilic/kube-start-stop/pkg/controller"
)

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, _ := clientcmd.BuildConfigFromFlags("", "")

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	cachedDiscovery := discocache.NewMemCacheClient(discoveryClient)
	restMapper := discovery.NewDeferredDiscoveryRESTMapper(cachedDiscovery, apimeta.InterfacesForUnstructured)
	restMapper.Reset()
	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	scaleClient, err := scale.NewForConfig(cfg, restMapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		fmt.Println(err)
		return
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	schedController := controller.NewController(kubeClient, scaleClient, exampleClient, kubeInformerFactory, exampleInformerFactory)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = schedController.Run(2, stopCh); err != nil {
		fmt.Println(err)
		return
	}
}
