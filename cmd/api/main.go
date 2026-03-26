package main

import (
	"ymmo/internal/api"
	// ... autres imports
)

func main() {
	// init DB, services, handlers...
	router := api.NewRouter(authH, propertyH, statsH, adminH, agentH)
	router.Run(":8080")
}
