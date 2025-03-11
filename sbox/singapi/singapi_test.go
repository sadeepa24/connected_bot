package singapi_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sadeepa24/connected_bot/sbox"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
)

func TestSing(t *testing.T) {
	instance, _, _ := singapi.NewsingAPI(context.Background(), "./sbox.json", nil, )
	err := instance.Start()

	if err != nil {
		log.Fatal(err)
	}

	// time.Sleep(2 * time.Minute)
	// ruid := "1c7a5143-bfeb-4cfd-b733-1f5e96edc949"
	uid, _ := uuid.NewV4()
	instance.AddUser(&sbox.Userconfig{
		Vlessgroup: &sbox.Vlessgroup{
			UUID: uid,
		},
		Inboundtag: "test_in",
		Usage:      0,
		Quota:      100,
		LoginLimit: 2,
	})
	go func() {
		time.Sleep(2 * time.Minute)
		instance.RemoveUser(&sbox.Userconfig{
			Vlessgroup: &sbox.Vlessgroup{
				UUID: uid,
			},
			Inboundtag: "test_in",
			Usage:      0,
			Quota:      100,
			LoginLimit: 2,
		})
	}()

	for {
		status, err := instance.GetstatusUser(&sbox.Userconfig{
			Vlessgroup: &sbox.Vlessgroup{
				UUID: uid,
			},
			Inboundtag: "test_in",
			Usage:      0,
			Quota:      100,
			LoginLimit: 2,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(status)
		time.Sleep(800 * time.Millisecond)
	}

	//time.Sleep(5 * time.Minute)

}
