package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/tools/clientcmd"
	student "xcodest.me/student/pkg/apis/student/v1"
	studentv1 "xcodest.me/student/pkg/generated/clientset/versioned/typed/student/v1"
	"xcodest.me/student/pkg/utils"
)

func main(){
	defaultConfig := os.ExpandEnv("$HOME/.kube/config")
	config , err := clientcmd.BuildConfigFromFlags("", defaultConfig)
	if err != nil {
		panic(err)
	}
	clientset, err := studentv1.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	watcher, err := clientset.Students("").Watch(context.Background(), metav1.ListOptions{})
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
			case watch.Modified:
				fmt.Printf("Student %s change\n%s\nn", student.Name, s)
			case watch.Deleted:
				fmt.Printf("Student %s deleted\n%s\n", student.Name, s)
		}
	}
}
