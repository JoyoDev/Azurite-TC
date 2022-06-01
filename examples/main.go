package main

import (
	"fmt"

	atc "github.com/joyodev/azurite-tc"
)

func main() {

	// default credentials for testing
	azuriteTC := atc.NewAzuriteTC("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")

	// create container
	azuriteTC.RunAzuriteContainer()

	// create test table and work with it
	azuriteTC.CreateTable("test")
	azuriteTC.UpdateTableValue("test", "user", "activeuser", "1")

	// read value
	value, err := azuriteTC.GetTableValue("test", "user", "activeuser")
	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Println(value)
	}

	// update again
	azuriteTC.UpdateTableValue("test", "user", "activeuser", "2")

	// read again
	value, err = azuriteTC.GetTableValue("test", "user", "activeuser")
	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Println(value)
	}

	// remove azurite container
	azuriteTC.RemoveAzuriteContainer()
}
