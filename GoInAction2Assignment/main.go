// Package main contains the main() func to run the application.
package main

import (
	server "GoInAction2Assignment/pkg/server"
)

// init will call server.InitServer() to initalise the application variables. This will only be called once
// in the duration of the application.
func init() {
	server.InitServer()
}

// main will call server.StartServer() to run the application.
func main() {

	server.StartServer()

}
