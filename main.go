package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"path/filepath"
	"terminal-ws/terminal"
)

var (
	restconfig *rest.Config
	kubeconfig *string
	client *kubernetes.Clientset
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func main() {
	flag.Parse()

	client = NewK8sClient()

	r := mux.NewRouter()

	r.PathPrefix("/api/sockjs/").Handler(terminal.CreateAttachHandler("/api/sockjs"))
	r.Handle("/pod/{namespace}/{pod}/shell", terminal.HandleExecShell(client, restconfig)).Methods("GET")
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"*"},
		AllowedOrigins: []string{"http://localhost:8081", "http://127.0.0.1:8000"},
		//AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	// Insert the middleware
	handler := c.Handler(r)
	log.Println("server run")
	log.Fatal(http.ListenAndServe(":8888", handler))
}

func NewK8sClient() *kubernetes.Clientset {
	var err error
	restconfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}