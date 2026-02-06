package main

import (

	"flag"
	"fmt"

	"net/http"
	"time"

	chaos "github.com/franciscocunha55/chaos-engineering/pkg/chaos"
	metrics "github.com/franciscocunha55/chaos-engineering/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"

)


func main() {

	// Flags
	intervalBetweenChaosTest := flag.Int("interval", 30, "Interval between chaos tests in seconds")
	chaosEngineeringNamespaceName := flag.String("namespace", "chaos-engineering-test", "Namespace for chaos engineering tests")
	flag.Parse()
	fmt.Printf("Interval between chaos tests: %d seconds\n", *intervalBetweenChaosTest)

	
	if err := metrics.Register(); err != nil {
		panic(fmt.Sprintf("Failed to register metrics: %s", err))
	}

	go func()  {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()


	clientSet, err := chaos.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	namespacesList := chaos.ListNamespaces(clientSet)

	chaos.CreateNamespace(clientSet, *chaosEngineeringNamespaceName, namespacesList)

	chaos.CreateDeployment(clientSet, *chaosEngineeringNamespaceName, "chaos-engineering-nginx")

	fmt.Printf("Starting chaos tests every %d seconds...\n", *intervalBetweenChaosTest)

	//ticker is a clock that ticks at regular intervals
	ticker := time.NewTicker(time.Duration(*intervalBetweenChaosTest) * time.Second)
	//Ensures the ticker is stopped when the program exits
	defer ticker.Stop()

	chaos.performChaosTest(clientSet, *chaosEngineeringNamespaceName)

	// ticker.C is a channel that receives a signal every time the ticker ticks, Repeat until program is terminated
	for range ticker.C {
        chaos.PerformChaosTest(clientSet, *chaosEngineeringNamespaceName)
    }
}
