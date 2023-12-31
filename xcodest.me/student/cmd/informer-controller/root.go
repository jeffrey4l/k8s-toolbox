package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	student "xcodest.me/student/pkg/apis/student/v1"
	st "xcodest.me/student/pkg/generated/clientset/versioned"
	studentClientset "xcodest.me/student/pkg/generated/clientset/versioned/typed/student/v1"
	stf "xcodest.me/student/pkg/generated/informers/externalversions"
	studentLister "xcodest.me/student/pkg/generated/listers/student/v1"
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

	ct2 := st.NewForConfigOrDie(config)
	clientset, err := studentClientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	listWatcher := cache.NewListWatchFromClient(clientset.RESTClient(), "students", "", fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	deletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	eventHandlerFuncs := cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
				deletedIndexer.Delete(obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.Add(key)
				deletedIndexer.Delete(newObj)
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

	factory := stf.NewSharedInformerFactory(ct2, 0)
	informer := factory.Xcodest().V1().Students()
	informer.Informer().AddEventHandler(eventHandlerFuncs)
	// indexer, informer := cache.NewIndexerInformer(listWatcher, &student.Student{}, 0, eventHandlerFuncs, cache.Indexers{})

	stopCh := make(chan struct{})
	go informer.Run(stopCh)
	defer close(stopCh)

	c := Controller{
		queue:          queue,
		indexer:        indexer,
		informer:       informer,
		clientset:      clientset,
		deletedIndexer: deletedIndexer,
	}
	go c.Run(stopCh)

	select {}
}

type Controller struct {
	queue          workqueue.RateLimitingInterface
	lister  *studentLister.StudentLister
	informer       cache.Controller
	clientset      *studentClientset.XcodestV1Client
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
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		obj, exists, err := c.deletedIndexer.GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			slog.Info(fmt.Sprintf("Student %s does not exists", key))
		}
		student, ok := obj.(*student.Student)
		if ok {
			slog.Info(fmt.Sprintf("Student %s is removed", student.GetName()))
			return nil
		}
		slog.Info(fmt.Sprintf("Get object is not student: %v", obj))
		return nil
	}

	student, ok := obj.(*student.Student)
	if !ok {
		slog.Info(fmt.Sprintf("Invalid student object: %v", obj))
		return nil
	}

	student.Status.Phase = "done"

	_, err = c.clientset.Students(student.Namespace).UpdateStatus(
		context.Background(), student, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Sync/Add/Update for Pod: %s", student.GetName()))
	return nil
}
