package builder

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWizardConf(t *testing.T) {

}

type Conec struct{}

func (c *Conec) Select(options []string, msg any) (string, error) {
	fmt.Println(msg)
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
	// fmt.Print("Enter a value: " + msg)
	// var choice string
	// fmt.Scanln(&choice)
	// return choice, nil
	fmt.Println(msg)
	return "<string from conec to "+msg+">", nil
}

func (c *Conec) AlertSend(msg string)  error {
	fmt.Println("ALERT:", msg)
	return nil
}


var conff = `{
 "dns": {
  "servers": [
   {
    "tag": "newdnssrv",
    "address": "1.1.1.1",
    "address_strategy": "prefer_ipv4",
    "address_fallback_delay": "300ms",
    "detour": "direct"
   }
  ],
  "rules": [
   {
    "inbound": "test",
    "action": "reject",
    "no_drop": true
   }
  ]
 },
 "inbounds": [
  {
   "type": "vless",
   "tag": "<<send the inbound tag name>>",
   "users": [
    {
     "name": "testname",
     "uuid": "<<send your' vless uuid>>"
    }
   ]
  }
 ],
 "route": {
  "rules": [
   {
    "inbound": "test",
    "auth_user": [
     "user1",
     "user2",
     "user3",
     "user4",
     "user5"
    ],
    "action": "reject",
    "no_drop": true
   }
  ],
  "final": "<<send the final tag name>>"
 },
 "experimental": {
  "clash_api": {
   "external_controller": "127.0.0.1"
  }
 }
}`


func TestWizardConfCreate(t *testing.T) {
	wsc := NewWizardConf([]byte(conff))
	
	// fmt.Println("St")
	// fmt.Println()
	// for i := range wsc{
	// 	fmt.Println(string(wsc[i]))
	// }

	fmt.Println("starting to export")
	buf := bytes.Buffer{}
	wsc.ExportTo(&buf, &Conec{})
	fmt.Println(buf.String())
		

}