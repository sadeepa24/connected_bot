package db_test

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/sadeepa24/connected_bot/db"
	sbConf "github.com/sadeepa24/connected_bot/sbox/conf"
)

func TestExportLink(t *testing.T) {
	testConf := db.Config{
		Id: 1,
		Name: "TestConfig",
		UUID: uuid.Nil.String(),
		Password: "testpassword",
	}
	testin := sbConf.Inboud{
		Id:            1,
		Name:          "TestInbound",
		Tag:           "test-tag",
		Type:          "trojan",
		Support:       []string{"support1", "support2"},
		ListenAddres:  "127.0.0.1",
		Listenport:    8080,
		Tlsenabled:    true,
		TransPortType: "ws",
		TransPortPath: "/test/path",
		Custom_info:   "custom-info",
		Domain:        "example.com",
		PublicIp:      "192.168.1.1",
	}


	testexpinfo := &sbConf.ExportInfo{
		Host: "test_host.com",
		Sni: "sni_test.com",
		Server: "",
	}

	fmt.Println(testConf.ExportUrlLink(testin, testexpinfo))
}