package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/client-go/tools/clientcmd"
    clientset "github.com/nevermosby/my-crd-controller/pkg/client/clientset/versioned"
    informers "github.com/nevermosby/my-crd-controller/pkg/client/informers/externalversions"
)

func main() {
    client, err := newCustomKubeClient()
    if err != nil {
        log.Fatalf("new kube client error: %v", err)
    }

    factory := informers.NewSharedInformerFactory(client, 30*time.Second)
    informer := factory.Mycontroller().V1alpha1().Websites()
    lister := informer.Lister()

    stopCh := make(chan struct{})
    factory.Start(stopCh)

    for {
        ret, err := lister.List(labels.Everything())
        if err != nil {
            log.Printf("list error: %v", err)
        } else {
            for _, website := range ret {
                log.Println(website)
            }
        }

        time.Sleep(5 * time.Second)
    }
}

func newCustomKubeClient() (clientset.Interface, error) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	
	//kubeconfig := flag.String("kubeconfig", "/Users/davidli/.kube/config", "kubeconfig file")


    config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create out-cluster kube cli configuration: %v", err)
    }

    cli, err := clientset.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create custom kube client: %v", err)
    }
    return cli, nil
}