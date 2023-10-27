package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/tools/clientcmd"
	student "xcodest.me/student/pkg/apis/student/v1"
	stclient "xcodest.me/student/pkg/generated/clientset/versioned"
	"xcodest.me/student/pkg/utils"
)

func main(){
	defaultConfig := os.ExpandEnv("$HOME/.kube/config")
	config , err := clientcmd.BuildConfigFromFlags("", defaultConfig)
	if err != nil {
		panic(err)
	}
	/*
	improt studentClientset "xcodest.me/student/pkg/generated/clientset/versioned/typed/student/v1"
	clientset, err := studentClientset.NewForConfig(config)
	watcher, err := clientset.Students("").Watch(context.Background(), metav1.ListOptions{})
	*/

	clientset := stclient.NewForConfigOrDie(config)

	watcher, err := clientset.XcodestV1().Students("").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan()  {
		student := event.Object.(*student.Student)
		s, err := utils.Obj2Json(student)
		if err != nil {
			panic(err)
		}
		s, err = utils.Obj2Yaml(student)
		if err != nil {
			panic(err)
		}
		switch event.Type {
			case watch.Added:
				fmt.Printf("Student %s added\n%s\n", student.Name, s)
				student.Status.Phase = "done"
				/*
				_, err := clientset.XcodestV1().Students(student.Namespace).Update(
					context.Background(), student, metav1.UpdateOptions{},
				)
				if err != nil {
					fmt.Printf("Get error during update: %s\n", err)
				}
				*/
				_, err = clientset.XcodestV1().Students(student.Namespace).UpdateStatus(
					context.Background(), student, metav1.UpdateOptions{},
				)
				if err != nil {
					fmt.Printf("Get error during update status: %s\n", err)
				}
			case watch.Modified:
				fmt.Printf("Student %s change\n%s\n", student.Name, s)
			case watch.Deleted:
				fmt.Printf("Student %s deleted\n%s\n", student.Name, s)
		}
	}
}
