package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/kevineaton/pregxas-api/api"
)

func main() {
	fmt.Println("\nStarting Pregxas API Server")
	rand.Seed(time.Now().UTC().UnixNano())

	r := api.SetupApp()

	api.Log("info", fmt.Sprintf("Listening on %v", api.Config.RootAPIPort), "server_start", map[string]string{
		"port": api.Config.RootAPIPort,
	})

	err := http.ListenAndServe(api.Config.RootAPIPort, r)
	if err != nil {
		api.Log("error", "Could not start server", "server_start_fail", map[string]string{
			"error": err.Error(),
		})
	}
}
