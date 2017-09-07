package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

type generic map[string]interface{}

func Evaluate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return
	}
	j, err := yaml.ToJSON(b)
	var deployment v1beta1.Deployment
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
		fmt.Fprintf(w, "%s\n", pi.Name)
	}
}

func main() {
	router := httprouter.New()
	router.POST("/evaluate", Evaluate)

	log.Fatal(http.ListenAndServe(":8080", router))
}
