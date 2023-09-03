package utils

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	"gopkg.in/yaml.v3"
)

func Obj2Json(obj interface{}) (string, error){
	objU, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
	if err != nil {
		panic(err)
	}

	objJson, err := json.MarshalIndent(objU, "", " ")
	if err != nil {
		panic(err)
	}

	objStr := string(objJson)
	return objStr, nil
}

func Obj2Yaml(obj interface{}) (string, error) {
	objU, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
	if err != nil {
		return "", err
	}

	objY, err := yaml.Marshal(objU)
	if err != nil {
		return "", err
	}
	return string(objY), nil
}
