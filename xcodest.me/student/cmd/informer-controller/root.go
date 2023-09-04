package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	student "xcodest.me/student/pkg/apis/student/v1"
	studentClientset "xcodest.me/student/pkg/generated/clientset/versioned/typed/student/v1"
	st "xcodest.me/student/pkg/generated/clientset/versioned"
)

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true})
	logger := slog.New(handler)
	slog.SetDefault(logger)


	defaultConfig := os.ExpandEnv("$HOME/.kube/config")
	config, err := clientcmd.BuildConfigFromFlags("", defaultConfig)
	if err != nil {
		panic(err)
	}

	st.NewForConfigOrDie(config).XcodestV1().Students()
	clientset, err := studentClientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	listWatcher := cache.NewListWatchFromClient(clientset.RESTClient(), "students", "", fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(
		listWatcher,
		&student.Student{},
		0,
		cache.ResourceEventHandlerDetailedFuncs{
			AddFunc: func(obj interface{}, isInInitialList bool) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
				} else {
					slog.Info(fmt.Sprintf("AddFunc error: %s", err))
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					queue.Add(key)
				} else {
					slog.Info(fmt.Sprintf("UpdateFunc error: %s", err))
				}
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
				} else {
					slog.Info(fmt.Sprintf("DeleteFunc error: %s", err))
				}
			},
		},
		cache.Indexers{})

	stopCh := make(chan struct{})
	go informer.Run(stopCh)
	defer close(stopCh)

	go Controller(stopCh, queue, indexer, informer, clientset)

	select {}
}

func Controller(stopCh chan struct{},
	queue workqueue.RateLimitingInterface,
	indexer cache.Indexer,
	informer cache.Controller,
	clientset *studentClientset.XcodestV1Client,
) {
	defer runtime.HandleCrash()
	defer queue.ShutDown()

	log.Println("Start the controller")
	for {
		log.Println("Try to get a new one")
		key, quit := queue.Get()
		log.Printf("Get %s\n", key)
		if quit {
			return
		}
		queue.Done(key)

		err := syncToStdout(key.(string), indexer, *clientset)
		if err != nil {
			slog.Error("error:", "error", err)
		}
	}
}

func syncToStdout(key string, indexer cache.Indexer, clientset studentClientset.XcodestV1Client) error {
	obj, exists, err := indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("Student %s not exists\n", key)
		return nil
	}

	student, ok := obj.(*student.Student)
	if !ok {
		log.Printf("Invalid student object: %v\n", obj)
		return nil
	}
	student.Status.Phase = "completed"

	// 更新
	slog.Info("update in namespace: ", "namespace", student.Namespace)
	_, err = clientset.Students(student.Namespace).UpdateStatus(context.Background(), student, metav1.UpdateOptions{})
	if err != nil {
		return err
	}


	log.Printf("Sync/Add/Update for Pod %s\n", student.GetName())
	return nil

}
