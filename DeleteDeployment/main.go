package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	//"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// Kubeconfig = flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "Kubernetes config file.")
	NS         = os.Getenv("NAMESPACES")
	Namespaces = strings.Split(NS, ",")
	DaysAgo    = os.Getenv("DAYS")
	LB1        = os.Getenv("LABEL1")
	LB2        = os.Getenv("LABEL2")
)

func ConnectKube() (clientset *kubernetes.Clientset) {

	// flag.Parse()
	// config, errc := clientcmd.BuildConfigFromFlags("", *Kubeconfig)
	// if errc != nil {
	// 	panic(errc.Error())
	// }
	// clientset, errC := kubernetes.NewForConfig(config)
	// if errC != nil {
	// 	panic(errC.Error())
	// }

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

func DeleteAll(ns, label string) {
	fmt.Println("Namespace, label: "+ns, label)
	clientset := ConnectKube()
	deletePolicy := metav1.DeletePropagationForeground

	fmt.Println("Deleting deployment...")
	deployment, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, d := range deployment.Items {
		if d.Labels["project"] == label {
			err := clientset.AppsV1().Deployments(ns).Delete(context.TODO(), d.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted deployment " + d.Name)
			}
		}
	}

	fmt.Println("Deleting ingress...")
	ingress, err := clientset.NetworkingV1beta1().Ingresses(ns).List(context.TODO(), metav1.ListOptions{})
	for _, i := range ingress.Items {
		if i.Labels["project"] == label {
			err := clientset.NetworkingV1beta1().Ingresses(ns).Delete(context.TODO(), i.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted ingress " + i.Name)
			}
		}
	}

	fmt.Println("Deleting configMap...")
	configMap, err := clientset.CoreV1().ConfigMaps(ns).List(context.TODO(), metav1.ListOptions{})
	for _, cm := range configMap.Items {
		if cm.Labels["project"] == label {
			err := clientset.CoreV1().ConfigMaps(ns).Delete(context.TODO(), cm.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted configMap " + cm.Name)
			}
		}
	}

	fmt.Println("Deleting secret...")
	secret, err := clientset.CoreV1().Secrets(ns).List(context.TODO(), metav1.ListOptions{})
	for _, s := range secret.Items {
		if s.Labels["project"] == label {
			err := clientset.CoreV1().Secrets(ns).Delete(context.TODO(), s.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted secret " + s.Name)
			}
		}
	}

	fmt.Println("Deleting service...")
	service, err := clientset.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
	for _, svc := range service.Items {
		if svc.Labels["project"] == label {
			err := clientset.CoreV1().Services(ns).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted service " + svc.Name)
			}
		}
	}

	fmt.Println("Deleting cronJob...")
	cronJob, err := clientset.BatchV1().CronJobs(ns).List(context.TODO(), metav1.ListOptions{})
	for _, cj := range cronJob.Items {
		if cj.Labels["project"] == label {
			err := clientset.BatchV1().CronJobs(ns).Delete(context.TODO(), cj.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted cronJob " + cj.Name)
			}
		}
	}

	fmt.Println("Deleting job...")
	job, err := clientset.BatchV1().Jobs(ns).List(context.TODO(), metav1.ListOptions{})
	for _, j := range job.Items {
		if j.Labels["project"] == label {
			err := clientset.BatchV1().Jobs(ns).Delete(context.TODO(), j.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted cronJob " + j.Name)
			}
		}
	}

	fmt.Println("Deleting pvc...")
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(ns).List(context.TODO(), metav1.ListOptions{})
	for _, p := range pvc.Items {
		if p.Labels["project"] == label {
			err := clientset.CoreV1().PersistentVolumeClaims(ns).Delete(context.TODO(), p.Name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Deleted pvc " + p.Name)
			}
		}
	}
	fmt.Println("******************************************")

}

func main() {
	fmt.Printf("Namespaces: %v\n", Namespaces)
	clientset := ConnectKube()
	timeNow := time.Now()
	for n, ns := range Namespaces {
		fmt.Println("##########################################")
		fmt.Println(n, " Namespace ", ns)

		deployment, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, d := range deployment.Items {
			if ((strings.Contains(d.Labels["project"], LB1)) || (strings.Contains(d.Labels["project"], LB2))) && (d.Labels["project"] == d.Name) {
				labelProject := d.Labels["project"]
				deploymentTime := d.Status.Conditions[1].LastUpdateTime
				daysDiff := timeNow.Sub(deploymentTime.Time).Hours() / 24
				fmt.Println("Label project: " + d.Labels["project"])
				fmt.Println(d.Name)
				fmt.Println(deploymentTime)
				fmt.Println(timeNow)
				fmt.Printf("Days diff: %f\n", daysDiff)
				DaysAgo, _ := strconv.ParseFloat(DaysAgo, 64)
				if DaysAgo < daysDiff {
					fmt.Println("TRUE")
					DeleteAll(ns, labelProject)
				}
				fmt.Println("------------------------")
			}
		}
	}
}
