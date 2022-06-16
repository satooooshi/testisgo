// Example showing how to patch Kubernetes resources.
// https://gist.github.com/dwmkerr/7332888e092156ce8ce4ea551b0c321f
// https://www.servicemesher.com/blog/manipulating-istio-and-other-custom-kubernetes-resources-in-golang/
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"fmt"
	"io/ioutil"
	"net/http"
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
	dr, err := client.Resource(drGVR).Namespace("istio-test").Patch(context.TODO(),
		drName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
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

func executeCanary(dynamicClient dynamic.Interface) {
	log.Print("start canary")
	//setVirtualServiceWeights(dynamicClient, virtualServiceName, weight1, weight2)
	fetchFromPrometheus(dynamicClient)
	log.Print("finish canary")
}

// Go構造体フィールド-jsonキーのマッピング規則
// https://dev.to/billylkc/parse-json-api-response-in-go-10ng
// 正解は[]byteにマッピングすれば良い
// https://qiita.com/chidakiyo/items/ac25449d49116ea189d0
type Resp struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name                                 string `json:"__name__"`
				App                                  string `json:"app"`
				Chart                                string `json:"chart"`
				ConnectionSecurityPolicy             string `json:"connection_security_policy"`
				DestinationApp                       string `json:"destination_app"`
				DestinationCanonicalRevision         string `json:"destination_canonical_revision"`
				DestinationCanonicalService          string `json:"destination_canonical_service"`
				DestinationCluster                   string `json:"destination_cluster"`
				DestinationPrincipal                 string `json:"destination_principal"`
				DestinationService                   string `json:"destination_service"`
				DestinationServiceName               string `json:"destination_service_name"`
				DestinationServiceNamespace          string `json:"destination_service_namespace"`
				DestinationVersion                   string `json:"destination_version"`
				DestinationWorkload                  string `json:"destination_workload"`
				DestinationWorkloadNamespace         string `json:"destination_workload_namespace"`
				Heritage                             string `json:"heritage"`
				InstallOperatorIstioIoOwningResource string `json:"install_operator_istio_io_owning_resource"`
				Instance                             string `json:"instance"`
				Istio                                string `json:"istio"`
				IstioIoRev                           string `json:"istio_io_rev"`
				Job                                  string `json:"job"`
				KubernetesNamespace                  string `json:"kubernetes_namespace"`
				KubernetesPodName                    string `json:"kubernetes_pod_name"`
				OperatorIstioIoComponent             string `json:"operator_istio_io_component"`
				PodTemplateHash                      string `json:"pod_template_hash"`
				Release                              string `json:"release"`
				Reporter                             string `json:"reporter"`
				RequestProtocol                      string `json:"request_protocol"`
				ResponseCode                         uint32 `json:"response_code,string"`
				ResponseFlags                        string `json:"response_flags"`
				ServiceIstioIoCanonicalName          string `json:"service_istio_io_canonical_name"`
				ServiceIstioIoCanonicalRevision      string `json:"service_istio_io_canonical_revision"`
				SidecarIstioIoInject                 string `json:"sidecar_istio_io_inject"`
				SourceApp                            string `json:"source_app"`
				SourceCanonicalRevision              string `json:"source_canonical_revision"`
				SourceCanonicalService               string `json:"source_canonical_service"`
				SourceCluster                        string `json:"source_cluster"`
				SourcePrincipal                      string `json:"source_principal"`
				SourceVersion                        string `json:"source_version"`
				SourceWorkload                       string `json:"source_workload"`
				SourceWorkloadNamespace              string `json:"source_workload_namespace"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// kubectl -n istio-system port-forward --address 0.0.0.0 $(kubectl -n istio-system get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090
func fetchFromPrometheus(dynamicClient dynamic.Interface) {
	//base, _ := url.Parse(`http://34.146.130.74:9090`)
	//reference, _ := url.Parse(`/api/v1/series?match[]=istio_requests_total{destination_service="reviews.istio-test.svc.cluster.local", destination_version="v2"}`)
	// reference, _ := url.Parse(`/api/v1/label/job/values`)
	//endpoint := base.ResolveReference(reference).String()
	// copied from brower
	//var svcname = "reviews"
	//var ns = "isito-test"
	//var version = "v2"
	req, _ := http.NewRequest("GET", `http://34.146.130.74:9090/api/v1/query?query=istio_requests_total%7Bdestination_service=%22reviews.istio-test.svc.cluster.local%22,%20destination_version=%22v2%22%7D`, nil)
	//req, _ := http.NewRequest("GET", `http://34.146.130.74:9090/api/v1/query?query=istio_requests_total%7Bdestination_service=%22`+svcname+`.`+ns+`.svc.cluster.local%22,%20destination_version=%22`+version+`%22%7D`, nil)
	//req, _ := http.NewRequest("GET", url.QueryEscape(`http://34.146.130.74:9090/api/v1/series?match[]=istio_requests_total{destination_service=reviews.istio-test.svc.cluster.local,destination_version=v2}`), nil)
	var client *http.Client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var newResp Resp
	// json.Unmarshalはjsonを構造体に変換します。
	if err := json.Unmarshal(body, &newResp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	fmt.Println("--------------------------------")
	for i, _ := range newResp.Data.Result {
		fmt.Printf("responseCode: %+v, value[1]: %+v\n", newResp.Data.Result[i].Metric.ResponseCode, newResp.Data.Result[i].Value[1])
	}

	limit := 5000 * time.Millisecond // loop 5 times
	begin := time.Now()
	var i = 0
	for now := range time.Tick(1000 * time.Millisecond) { // every 1000ms
		fmt.Println("Tick!!")
		req, _ := http.NewRequest("GET", `http://34.146.130.74:9090/api/v1/query?query=istio_requests_total%7Bdestination_service=%22reviews.istio-test.svc.cluster.local%22,%20destination_version=%22v2%22%7D`, nil)
		//req, _ := http.NewRequest("GET", url.QueryEscape(`http://34.146.130.74:9090/api/v1/query?query=istio_requests_total{destination_service="reviews.istio-test.svc.cluster.local",destination_version="v2"}`), nil)
		var client *http.Client = &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var newResp Resp
		// json.Unmarshalはjsonを構造体に変換します。
		if err := json.Unmarshal(body, &newResp); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("--------------------------------")
		for i, _ := range newResp.Data.Result {
			fmt.Printf("responseCode: %+v, value[1]: %+v\n", newResp.Data.Result[i].Metric.ResponseCode, newResp.Data.Result[i].Value[1])
		}

		var weight1 [3]uint32 = [3]uint32{20, 70, 90}
		var weight2 [3]uint32 = [3]uint32{70, 20, 0}
		setVirtualServiceWeights(dynamicClient, "reviews", weight1[i], weight2[i])

		// time.Tickで取得した現在時間とループ開始直前の時間の差分でループを止めるか決める
		if now.Sub(begin) >= limit {
			break
		}

		if i == 2 {
			break
		}
		i++
	}

	// Error Rate の閾値は 0.1% ( これを越えた場合 rollback する ) do fetchFromPrometheus() else rollback
	// もし Error Rate 等がある閾値を越えてしまった場合は、自動的に"すべての"トラフィックを現在の安定版 “stable” に戻し、canary を停止します (= rollback) 。
	// manually for i /api/catalog/hello 50 times
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
	// setDestinationRuleCb(dynamicClient, "catalog", 3)

	//executeCanary(dynamicClient)
	// export KUBECONFIG='/root/.kube/config' && export NAMESPACE='default' && go run k8s-patch-virtualservice.go
	listIngress(dynamicClient)
}

func listIngress(client dynamic.Interface) error {
	//  Create a GVR which represents an Istio Virtual Service.
	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingress",
	}

	//  Apply the patch to the 'service2' service.
	res, err := client.Resource(gvr).Namespace("default").List(context.TODO(), metav1.ListOptions{})
	log.Print(res)
	if err != nil {
		log.Print(err)
	}
	return err
}

func addRouteToVs(client dynamic.Interface, ns string, virtualServiceName string, host string, port uint32) error {
	//  Create a GVR which represents an Istio Virtual Service.
	gvr := schema.GroupVersionResource{
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
	vs, err := client.Resource(gvr).Namespace(ns).Patch(context.TODO(), virtualServiceName, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	log.Print(vs)
	if err != nil {
		log.Print(err)
	}
	return err
}
