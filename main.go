package main

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	deleteOptions := metav1.NewDeleteOptions(0);
	fmt.Printf("%+v{}\n", deleteOptions);
}
