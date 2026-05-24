package main

import "fmt"

func printServeStarted(addr string) {
	fmt.Printf("PreDeploy Guard server listening on http://%s\n", addr)
}
