// Example showing how to patch Kubernetes resources.
// https://gist.github.com/dwmkerr/7332888e092156ce8ce4ea551b0c321f
// https://www.servicemesher.com/blog/manipulating-istio-and-other-custom-kubernetes-resources-in-golang/
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	//  Leave blank for the default context in your kube config.
	//context = ""

	//  Name of the VirtualService to weight, and the two weight values.
	// total destination weight MUST be 100
	virtualServiceName = "reviews"
	weight1            = uint32(99)
	weight2            = uint32(1)
)

//  patchStringValue specifies a patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

//  patchStringValue specifies a patch operation for a uint32.
type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

func setVirtualServiceWeights(client dynamic.Interface, virtualServiceName string, weight1 uint32, weight2 uint32) error {
	//  Create a GVR which represents an Istio Virtual Service.
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}

	//  Weight the two routes - 50/50.
	patchPayload := make([]patchUInt32Value, 2)
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/http/0/route/0/weight"
	patchPayload[0].Value = weight1
	patchPayload[1].Op = "replace"
	patchPayload[1].Path = "/spec/http/0/route/1/weight"
	patchPayload[1].Value = weight2
	patchBytes, _ := json.Marshal(patchPayload)

	//  Apply the patch to the 'service2' service.
	vs, err := client.Resource(virtualServiceGVR).Namespace("istio-test").Patch(context.TODO(), virtualServiceName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	log.Print(vs)
	if err != nil {
		log.Print(err)
	}
	return err
}

func setDestinationRuleLb(client dynamic.Interface, drName string, newLb string) error {
	//  Create a GVR which represents an Istio Virtual Service.
	drGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "destinationrules",
	}

	//  Weight the two routes - 50/50.
	patchPayload := make([]patchStringValue, 1)
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/trafficPolicy/loadBalancer/simple"
	patchPayload[0].Value = newLb
	patchBytes, _ := json.Marshal(patchPayload)

	//  Apply the patch to the 'service2' service.
	dr, err := client.Resource(drGVR).Namespace("istio-test").Patch(context.TODO(), drName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	log.Print(dr)
	if err != nil {
		log.Print(err)
	}
	return err
}

// Configuring the circuit breaker
func setDestinationRuleCb(client dynamic.Interface, drName string, newCnt uint32) error {
	//  Create a GVR which represents an Istio Virtual Service.
	drGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "destinationrules",
	}

	//  Weight the two routes - 50/50.
	patchPayload := make([]patchUInt32Value, 3)
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/trafficPolicy/connectionPool/tcp/maxConnections"
	patchPayload[0].Value = newCnt
	patchPayload[1].Op = "replace"
	patchPayload[1].Path = "/spec/trafficPolicy/connectionPool/http/http1MaxPendingRequests"
	patchPayload[1].Value = newCnt
	patchPayload[2].Op = "replace"
	patchPayload[2].Path = "/spec/trafficPolicy/connectionPool/http/maxRequestsPerConnection"
	patchPayload[2].Value = newCnt
	patchBytes, _ := json.Marshal(patchPayload)

	//  Apply the patch to the 'service2' service.
	dr, err := client.Resource(drGVR).Namespace("istio-test").Patch(context.TODO(), drName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	log.Print(dr)
	if err != nil {
		log.Print(err)
	}
	return err
}

func main() {
	kubeconfig := os.Getenv("KUBECONFIG") // os.GEtenv gets environment variable
	namespace := os.Getenv("NAMESPACE")

	if len(kubeconfig) == 0 || len(namespace) == 0 {
		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	// Creates the dynamic interface.
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	//  Re-balance the weights of the hosts in the virtual service.
	//setVirtualServiceWeights(dynamicClient, virtualServiceName, weight1, weight2)
	//setDestinationRuleLb(dynamicClient, "catalog", "ROUND_ROBIN") //LEAST_CONN RANDOM ROUND_ROBIN
	setDestinationRuleCb(dynamicClient, "catalog", 3)
}
