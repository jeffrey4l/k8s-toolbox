package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	clientset "xcodest.me/student/pkg/generated/clientset/versioned"
	informers "xcodest.me/student/pkg/generated/informers/externalversions"
	studentV1Lister "xcodest.me/student/pkg/generated/listers/student/v1"
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

	cs := clientset.NewForConfigOrDie(config)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	deletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	eventHandlerFuncs := cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				deletedIndexer.Add(obj)
				queue.Add(key)
			}
		},
	}

	factory := informers.NewSharedInformerFactory(cs, 0)
	informer := factory.Xcodest().V1().Students()
	informer.Informer().AddEventHandler(eventHandlerFuncs)
	slog.Info(fmt.Sprintf("indexer: %#v", informer.Informer().GetIndexer().GetIndexers()))
	return

	stopCh := make(chan struct{})
	go informer.Informer().Run(stopCh)
	defer close(stopCh)

	c := Controller{
		queue:          queue,
		lister:         informer.Lister(),
		informer:       informer.Informer(),
		clientset:      cs,
		deletedIndexer: deletedIndexer,
	}
	go c.Run(stopCh)

	select {}
}

type Controller struct {
	queue          workqueue.RateLimitingInterface
	lister         studentV1Lister.StudentLister
	informer       cache.SharedIndexInformer
	clientset      *clientset.Clientset
	deletedIndexer cache.Indexer
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	slog.Info("Start the controller")
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(errors.New("Timed out wait for cache to sync"))
	}
	for {
		slog.Debug("Try to get a new one")
		key, quit := c.queue.Get()
		if quit {
			return
		}
		err := c.syncToStdout(key.(string))
		if err != nil {
			slog.Error("Get error", "error", err)
		}
		c.queue.Done(key)
	}
}

func (c *Controller) syncToStdout(key string) error {
	namespace, name, _ := cache.SplitMetaNamespaceKey(key)
	student, err := c.lister.Students(namespace).Get(name)
	if err != nil {
		obj, exists, err := c.deletedIndexer.GetByKey(key)
		slog.Info(fmt.Sprintf("%s, %s", obj, exists))
		return err
	}
	student.Status.Phase = "done"

	_, err = c.clientset.XcodestV1().Students(student.Namespace).UpdateStatus(
		context.Background(), student, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Sync/Add/Update for Pod: %s", student.GetName()))
	return nil
}
