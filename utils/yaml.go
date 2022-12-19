package utils

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Property struct {
	Name       string
	Desc       string
	DefaultVal string
}

type contain struct {
	image           string
	imagePullPolicy string
	name            string
}

func ChangeYaml(sourceFilePath string, outputFilePath string, changeMap map[string]interface{}) error {
	filebytes, err := ioutil.ReadFile(sourceFilePath)
	if err != nil {
		return err
	}
	result := make(map[interface{}]interface{})
	err = yaml.Unmarshal(filebytes, &result)
	if err != nil {
		return err
	}
	mapChange(result, changeMap)
	text, err := yaml.Marshal(&result)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFilePath, text, 0777)
}

func mapChange(sourceMap map[interface{}]interface{}, changeMap map[string]interface{}) {
	if sourceMap == nil || changeMap == nil {
		return
	}
	for k, v := range changeMap {
		if _, ok := sourceMap[k]; ok {
			val, isStr := v.(string)
			if isStr == true {
				sourceMap[k] = val
				fmt.Println("Change ", k, " ", val)
				continue
			} else {
				x := sourceMap[k].(map[interface{}]interface{})
				y := changeMap[k].(map[string]interface{})
				mapChange(x, y)
			}
		} else {
			sourceMap[k] = changeMap[k]
			fmt.Println("Change ", k, " ", changeMap[k])
		}
	}
}

func GeneratePodYaml(nameSpace string, PodName string, isHostNetwork bool) (string, error) {
	ctn := make([]map[string]string, 0, 1)
	ctn = append(ctn, map[string]string{
		"image":           PodName,
		"imagePullPolicy": "IfNotPresent",
		"name":            PodName,
	})

	mp := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"k8s-app": PodName + "-app",
			},
			"name":      PodName,
			"namespace": nameSpace,
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"k8s-app": PodName + "-app",
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"k8s-app": PodName + "-app",
					},
				},
				"spec": map[string]interface{}{
					"hostNetwork": isHostNetwork,
					"nodeSelector": map[string]interface{}{
						"node-role.kubernetes.io/master": "",
					},
					"containers":    ctn,
					"restartPolicy": "Always",
				},
			},
		},
	}
	text, err := yaml.Marshal(&mp)
	if err != nil {
		return "", err
	}
	fileName := PodYamlPrefix + PodName + ".yaml"
	err = ioutil.WriteFile(fileName, text, 0666)
	return fileName, nil
}
