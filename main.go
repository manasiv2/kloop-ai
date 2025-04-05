package main

import (
	"fmt"

	"github.com/manasiv2/kloop-ai/kube"
)

func main() {
	errors := kube.CheckPodPhases("") // or "default" for just one namespace
	fmt.Println(len(errors))

	for _, err := range errors {
		fmt.Println(err.Summary) // or print full struct if you want
	}
}
