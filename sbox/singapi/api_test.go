package singapi_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/sbox/singapi"
	"github.com/sagernet/sing-box/option"
	"go.uber.org/zap"
)

type testItem[T any] struct {
	expecterr bool
	item      T
}

//TODO: later 😒
var testopts = option.Options{
	Log: &option.LogOptions{
		Disabled: true,
	},

	Inbounds: []option.Inbound{},
	Outbounds: []option.Outbound{},

	Route: &option.RouteOptions{
		
		Rules: []option.Rule{
			option.Rule{},
			option.Rule{},
			option.Rule{},
		},
		Final: "direct",
	},
}




func TestSingApi(t *testing.T) {
	dLogger, _ := zap.NewDevelopment()

	boxapi, opts, err := singapi.NewsingAPI(context.Background(), "./config.json", dLogger)
	if err != nil {
		t.Error(err.Error())
	}
	_ = opts
	err = boxapi.Start()
	if err != nil {
		t.Error(err.Error())
	}
	
	testUsers := []testItem[db.Config]{
		{
			item: db.Config{
				Id:         22,
				Name:       "User1",
				UUID: "",
				Password:   "password1",
				Active:     true,
				UserID:     101,
				OutboundID: 1,
				InboundIds: db.InIds{1, 2},
				Usage:      1000,
				Download:   500,
				Upload:     500,
				Quota:      2000,
				LoginLimit: 3,
			},

			
		},
		{
			item: db.Config{
				Id:         23,
				Name:       "User2",
				Password:   "password2",
				Active:     true,
				UserID:     102,
				OutboundID: 2,
				InboundIds: db.InIds{1, 2},
				Usage:      2000,
				Download:   1000,
				Upload:     1000,
				Quota:      3000,
				LoginLimit: 3,
			},
		},
		{
			item: db.Config{
				Id:         24,
				Name:       "User3",
				Password:   "password3",
				Active:     true,
				UserID:     103,
				OutboundID: 2,
				InboundIds: db.InIds{2},
				Usage:      3000,
				Download:   1500,
				Upload:     1500,
				Quota:      4000,
				LoginLimit: 3,
			},
		},
		{
			expecterr: true,
			item: db.Config{Id:         25,
				Name:       "User4",
				Password:   "password4",
				Active:     true,
				UserID:     104,
				OutboundID: 2,
				InboundIds: db.InIds{10},
				Usage:      4000,
				Download:   2000,
				Upload:     2000,
				Quota:      5000,
				LoginLimit: 3,},
		},
		{
			item: db.Config{Id:         26,
				Name:       "User5",
				Password:   "password5",
				Active:     true,
				UserID:     105,
				OutboundID: 1,
				InboundIds: db.InIds{1},
				Usage:      5000,
				Download:   2500,
				Upload:     2500,
				Quota:      6000,
				LoginLimit: 3,},
		},
		{
			expecterr: true,
			item: db.Config{
				Id:         27,
				Name:       "User6",
				UUID:       "nouuid",
				Password:   "password6",
				Active:     true,
				UserID:     106,
				OutboundID: 2,
				InboundIds: db.InIds{2},
				Usage:      6000,
				Download:   3000,
				Upload:     3000,
				Quota:      7000,
				LoginLimit: 3,
			},
		},
		{
			item: db.Config{
				Id:         28,
				Name:       "User7",
				Password:   "password7",
				Active:     true,
				UserID:     107,
				OutboundID: 1,
				InboundIds: db.InIds{2},
				Usage:      7000,
				Download:   3500,
				Upload:     3500,
				Quota:      8000,
				LoginLimit: 3,
			},
		},
		{
			item: db.Config{
				Id:         29,
				Name:       "User8",
				Password:   "password8",
				Active:     true,
				UserID:     108,
				OutboundID: 2,
				InboundIds: db.InIds{1},
				Usage:      8000,
				Download:   4000,
				Upload:     4000,
				Quota:      9000,
				LoginLimit: 3,
			},
		},
		{
			expecterr: true,
			item: db.Config{
				Id:         30,
				Name:       "User9",
				Password:   "password9",
				Active:     true,
				UserID:     109,
				OutboundID: 2,
				InboundIds: db.InIds{1},
				Usage:      9000,
				Download:   4500,
				Upload:     4500,
				Quota:      10000,
				LoginLimit: 0,
			},
		},
		{
			item: db.Config{
				Id:         31,
				Name:       "User10",
				Password:   "password10",
				Active:     true,
				UserID:     110,
				OutboundID: 1,
				InboundIds: db.InIds{1},
				Usage:      10000,
				Download:   5000,
				Upload:     5000,
				Quota:      11000,
				LoginLimit: 3,
			},
		},
	}

	for i, user := range testUsers {
		if user.item.UUID != "nouuid" {
			newuuid, err := uuid.NewV4()
			if err != nil {
				user.expecterr = true
			} else {
				testUsers[i].item.UUID = newuuid.String()
			}
		
		}
		// dLogger.Info("addding user " + user.item.Name)
		_, err = boxapi.AddConfig(&testUsers[i].item)
		if err != nil && !user.expecterr {
			t.Error("user adding failed didn't expect err but got", err.Error())
			if boxerr, ok := err.(singapi.Error); ok {
				t.Log("isboxerr:", boxerr.IsBoxErr())
			}
		}
		if user.expecterr && err == nil {
			t.Log(i, user.item.Name)
			t.Log("error expected but got no error")
		}
	}


	for i, user := range testUsers {

		_, err = boxapi.GetStatusConfig(&user.item)
		
		if err != nil && !user.expecterr {
			t.Error("user adding failed didn't expect err but got", err.Error(), user.item.Name)
			if boxerr, ok := err.(singapi.Error); ok {
				t.Log("isboxerr:", boxerr.IsBoxErr())
			}
		}
		if user.expecterr && err == nil {
			t.Log(i, user.item.Name)
			t.Log("error expected but got no error")
		}

	}
	for i, user := range testUsers {


		_, err = boxapi.AddConfigReset(&user.item)
		
		if err != nil && !user.expecterr {
			t.Error("user adding failed didn't expect err but got", err.Error())
			if boxerr, ok := err.(singapi.Error); ok {
				t.Log("isboxerr:", boxerr.IsBoxErr())
			}
		}
		if user.expecterr && err == nil {
			t.Log(i, user.item.Name)
			t.Log("error expected but got no error")
		}

	}


}