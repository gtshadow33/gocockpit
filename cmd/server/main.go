package main

import (
	"fmt"

	"gocockpit/internal/auth"
)

func main() {
	var username string
	var password string

	fmt.Print("Usuario: ")
	fmt.Scanln(&username)

	fmt.Print("Contraseña: ")
	fmt.Scanln(&password)

	err := auth.Authenticate(username, password)

	if err != nil {
		fmt.Println("❌ Autenticación fallida:", err)
		return
	}

	fmt.Println("✅ Autenticación correcta")
}