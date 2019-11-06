
./vendor/k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
github.com/nevermosby/my-crd-controller/pkg/client \
github.com/nevermosby/my-crd-controller/pkg/apis "mycontroller:v1alpha1" \
--go-header-file /Users/davidli/gh/myk8scrd/src/github.com/nevermosby/my-crd-controller/hack/custom-boilerplate.go.txt