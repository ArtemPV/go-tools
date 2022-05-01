package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/julienschmidt/httprouter"

	"html/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Image struct {
	Namespaces  string
	Deployments []string
	URLs        []string
	Tags        string
}

var (
	// Kubeconfig = flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "Kubernetes config file.")
	NS         = os.Getenv("NAMESPACES")
	LB1        = os.Getenv("LABEL1")
	LB2        = os.Getenv("LABEL2")
	Namespaces = strings.Split(NS, ",")
)

func ConnectKube() (clientset *kubernetes.Clientset) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return
}

func GetAllImages() (images []Image, err error) {
	images = []Image{}
	deploymentList := []string{}
	ingressList := []string{}
	fmt.Println(Namespaces)

	clientset := ConnectKube()

	for n, ns := range Namespaces {
		fmt.Println(n, " - ", ns)
		labelProject := ""
		ingressLabel := ""
		deployment, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		ingress, err := clientset.ExtensionsV1beta1().Ingresses(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, d := range deployment.Items {
			if !(strings.Contains(d.Labels["project"], LB1)) && !(strings.Contains(d.Labels["project"], LB2)) {
				break
			}
			if labelProject == "" {
				labelProject = d.Labels["project"]
			}
			if labelProject != d.Labels["project"] {
				xz := Image{ns, deploymentList, ingressList, labelProject}
				images = append(images, xz)
				deploymentList = nil
				ingressList = nil
				labelProject = d.Labels["project"]
				deploymentList = append(deploymentList, d.Name)
			} else {
				deploymentList = append(deploymentList, d.Name)
			}
			if ingressLabel != labelProject {
				ingressLabel = labelProject
				for _, i := range ingress.Items {
					if i.Labels["project"] == d.Labels["project"] {
						for host := range i.Spec.Rules {
							ingressList = append(ingressList, i.Spec.Rules[host].Host)
						}
					}
				}
			}
		}
		xz := Image{ns, deploymentList, ingressList, labelProject}
		images = append(images, xz)
		deploymentList = nil
		ingressList = nil
	}
	return
}

func GetImages(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	images, err := GetAllImages()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	main := filepath.Join("public", "html", "imagesDynamicPage.html")
	common := filepath.Join("public", "html", "common.html")

	tmpl, err := template.ParseFiles(main, common)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "images", images)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
}

func main() {

	r := httprouter.New()
	routes(r)

	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal(err)
	}
}

func routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))
	r.GET("/", GetImages)
}
