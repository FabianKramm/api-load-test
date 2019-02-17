package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	testRequestsPerSecond()
}

func testConcurrentConnections() {
	connections := make(chan int)
	fmt.Println("Start requesting")

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err)
	}

	// Up to 1 million connections
	for i := 0; i < 10000; i++ {
		go func() {
			client, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}

			iface, err := client.Core().Pods(metav1.NamespaceSystem).Watch(metav1.ListOptions{})
			if err != nil {
				log.Println(err)
				return
			}

			connections <- 1
			resultChan := iface.ResultChan()

			for event := range resultChan {
				// Do noop operation
				event.Object.GetObjectKind()
			}

			connections <- -1
		}()
	}

	numConnections := 0
	for connection := range connections {
		numConnections += connection

		fmt.Printf("\rOpened connections: %d", numConnections)
	}
}

func testRequestsPerSecond() {
	requests := make(chan int)
	now := time.Now()

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Start requesting")

	// Start 1000 go routines
	for i := 0; i < 1000; i++ {
		go func() {
			client, err := kubernetes.NewForConfig(config)
			if err != nil {
				panic(err)
			}

			for true {
				_, err := client.Core().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{})
				if err != nil {
					log.Println(err)
					continue
				}

				requests <- 1
			}
		}()
	}

	numRequests := 0
	for request := range requests {
		numRequests += request

		fmt.Printf("\rRequests: %f req/s", float64(numRequests)/time.Now().Sub(now).Seconds())
	}
}
