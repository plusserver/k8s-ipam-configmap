package main

import (
	"flag"
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/Nexinto/go-ipam"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Nexinto/k8s-ipam-shared"

	ipamclientset "github.com/Nexinto/k8s-ipam/pkg/client/clientset/versioned"
)

func main() {

	flag.Parse()

	// If this is not set, glog tries to log into something below /tmp which doesn't exist.
	flag.Lookup("log_dir").Value.Set("/")

	if e := os.Getenv("LOG_LEVEL"); e != "" {
		if l, err := log.ParseLevel(e); err == nil {
			log.SetLevel(l)
		} else {
			log.SetLevel(log.WarnLevel)
			log.Warnf("unkown log level %s, setting to 'warn'", e)
		}
	}

	var kubeconfig string

	if e := os.Getenv("KUBECONFIG"); e != "" {
		kubeconfig = e
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	kubernetes, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err.Error())
	}

	ipamclient, err := ipamclientset.NewForConfig(clientConfig)
	if err != nil {
		panic(err.Error())
	}

	var network string

	if e := os.Getenv("IPAM_NETWORK"); e != "" {
		network = e
	} else {
		panic("need IPAM_NETWORK")
	}

	nameTemplate, err := MakeNameTemplate()
	if err != nil {
		panic(err)
	}

	am, err := ipam.NewConfigMapIpam(kubernetes, network)
	if err != nil {
		panic(err.Error())
	}

	c := &Controller{
		Kubernetes: kubernetes,
		IpamClient: ipamclient,
		SharedController: ipamshared.SharedController{
			Kubernetes:   kubernetes,
			IpamClient:   ipamclient,
			Ipam:         am,
			Tag:          "ConfigMap",
			NameTemplate: nameTemplate,
			IpamName:     "ConfigMap",
		},
	}

	c.Initialize()
	c.Start()
}

func MakeNameTemplate() (nameTemplate *template.Template, err error) {
	if e := os.Getenv("NAME_TEMPLATE"); e != "" {
		nameTemplate, err = template.New("name").Parse(e)
	} else {
		nameTemplate, err = template.New("name").Parse("{{.Tag}}.{{.Namespace}}.{{.Name}}")
	}
	return
}
