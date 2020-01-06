# Kubernetes CRD & controller for building a website in one step

## Goal

Inspired by the implementation of [website controller](https://github.com/luksa/k8s-website-controller/) of [Kubernetes in Action ](https://github.com/luksa/kubernetes-in-action), You can easily build up a website via a git repo and nginx /w kubernetes workload.

You can find more via the blog: 
- 中文版：[https://davidlovezoe.club/wordpress/archives/690](https://davidlovezoe.club/wordpress/archives/690)
- English: TODO

## Development
1. Define your own crd(CustomResourceDefinitions) struct, for example:
    ```go
    type Website struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   WebsiteSpec   `json:"spec"`
        Status WebsiteStatus `json:"status"`
    }

    type WebsiteSpec struct {
        GitRepo        string `json:"gitRepo"`
        DeploymentName string `json:"deploymentName"`
        Replicas       *int32 `json:"replicas"`
    }

    type WebsiteStatus struct {
        AvailableReplicas int32 `json:"availableReplicas"`
    }

    type WebsiteList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`

        Items []Website `json:"items"`
    }
    ```
2. Generate the controller stuff, not write it

    The kubernetes community provides serveral ways to automatically generate the controller stuff, like:
    - [Code generator](https://github.com/kubernetes/code-generator)
    - [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
    - [Operator-sdk](https://github.com/operator-framework/operator-sdk)

    I use `code generator` for this project and wrap it as a shell script *`codegen.sh`*:

    ```bash
    ./vendor/k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
    github.com/nevermosby/my-crd-controller/pkg/client \
    github.com/nevermosby/my-crd-controller/pkg/apis "mycontroller:v1alpha1" \
    --go-header-file /Users/davidli/gh/myk8scrd/src/github.com/nevermosby/my-crd-controller/hack/custom-boilerplate.go.txt
    ```

3. Write your event handler
   - Create the nginx deployment based on the git repo
   - Create the nodeport service based on the nginx deployment

## Deploy the CRD and controller

1. Deploy the CRD via YAML
2. Deploy the controller via deployment