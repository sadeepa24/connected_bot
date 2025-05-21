package main

import (
	"fmt"
)

//dummy interface for testing
type Conec struct{}

func (c *Conec) Select(options []string, msg string) (string, error) {
	fmt.Println(msg)
	fmt.Println("Select an option:\n")
	for i, option := range options {
		fmt.Printf("%d: %s \n", i+1, option)
	}

	var choice int
	fmt.Print("\nEnter choice: ")
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(options) {
		fmt.Println("Invalid input.")
		return "", nil
	}

	return options[choice-1], nil
}

func (c *Conec) ReciveVal(msg string) (string, error) {
	fmt.Print("Enter a value: " + msg)
	var choice string
	fmt.Scanln(&choice)
	return choice, nil
}

func (c *Conec) AlertSend(msg string)  error {
	fmt.Println("ALERT:", msg)
	return nil
}



func main() {
	// b := builder.NewBuilder(&Conec{})
	// //b.RouteFieldChange()
	// b.DNSfieldChange()
	// b.OutboundFieldsChange("sock")
}