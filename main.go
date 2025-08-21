/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/remiehneppo/material-management/cmd"

// @title Materials Management API
// @version 1.0
// @description Materials Management API with Golang
// @termsOfService http://swagger.io/terms/

// @contact.name Bao Tran
// @contact.url http://github.com/remiehneppo
// @contact.email bao.tran080898@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8088
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cmd.Execute()
}
