package main

import (
	"log"
	"net/http"
	"path/filepath"
	"terminal-ws/terminal"

	flag "github.com/spf13/pflag"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	restconfig     *rest.Config
	kubeconfig     string
	client         *kubernetes.Clientset
	allowedOrigins []string
	port           string
)

func init() {
	flag.StringVar(&port, "port", "80", "serve port")
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "absolute path to the kubeconfig file")
	flag.StringSliceVar(&allowedOrigins, "allowedOrigins", []string{"*"}, "cors allowedOrigins")
}

func main() {
	flag.Parse()

	client = newK8sClient()

	r := mux.NewRouter()

	r.PathPrefix("/api/sockjs/").Handler(terminal.CreateAttachHandler("/api/sockjs"))
	r.Handle("/pod/{namespace}/{pod}/shell", terminal.HandleExecShell(client, restconfig)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		// Debug: true,
	})

	// Insert the middleware
	log.Println("allowedOrigins: ", allowedOrigins)
	log.Println("kubeconfig: ", kubeconfig)
	log.Println("server run at port: " + port)

	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}

func newK8sClient() *kubernetes.Clientset {
	var err error
	restconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
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
