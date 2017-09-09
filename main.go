package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/julienschmidt/httprouter"
	yaml_util "k8s.io/apimachinery/pkg/util/yaml"
	core "k8s.io/client-go/pkg/api/v1"
	ar_v1alpha1 "k8s.io/client-go/pkg/apis/admissionregistration/v1alpha1"
	apps_v1beta1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
)

type deploymentInitializer func(d *apps_v1beta1.Deployment) error

var initContainersInitializer = func(d *apps_v1beta1.Deployment) error {
	initContainer := core.Container{
		Name:  "init",
		Image: "init",
	}

	d.Spec.Template.Spec.InitContainers = append(d.Spec.Template.Spec.InitContainers, initContainer)
	return nil
}

var volumeInitializer = func(d *apps_v1beta1.Deployment) error {
	volume := core.Volume{
		Name: "test",
	}

	d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, volume)
	return nil
}

func removePendingInitializer(d *apps_v1beta1.Deployment, name string) {
	is := d.ObjectMeta.Initializers.Pending[:0]
	for _, x := range d.ObjectMeta.Initializers.Pending {
		if x.Name != name {
			is = append(is, x)
		}
	}
	d.ObjectMeta.Initializers.Pending = is
}

var registeredInitializers = map[string]deploymentInitializer{
	"volume.kisc.kubernetes.io":         volumeInitializer,
	"init-container.kisc.kubernetes.io": initContainersInitializer,
}

var registeredRules = struct {
	sync.RWMutex
	m map[string][]ar_v1alpha1.Rule
}{m: make(map[string][]ar_v1alpha1.Rule)}

func Evaluate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}
	j, err := yaml_util.ToJSON(b)
	var deployment apps_v1beta1.Deployment
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(j, &deployment)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}

	for _, pi := range deployment.ObjectMeta.Initializers.Pending {
		if f, ok := registeredInitializers[pi.Name]; ok {
			err = f(&deployment)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
				return
			} else {
				removePendingInitializer(&deployment, pi.Name)
			}
		}
	}

	out, err := yaml.Marshal(deployment)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", out)
}

func Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}
	j, err := yaml_util.ToJSON(b)
	var configuration ar_v1alpha1.InitializerConfiguration
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(j, &configuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}

	for _, i := range configuration.Initializers {
		for _, r := range i.Rules {
			registeredRules.Lock()
			registeredRules.m[i.Name] = append(registeredRules.m[i.Name], r)
			registeredRules.Unlock()
		}
	}

	fmt.Fprintf(w, "%+v\n", registeredRules.m)
}

func main() {
	router := httprouter.New()
	router.POST("/evaluate", Evaluate)
	router.POST("/register", Register)

	log.Fatal(http.ListenAndServe(":8080", router))
}
