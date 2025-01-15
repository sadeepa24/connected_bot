package db_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	sdb "github.com/sadeepa24/connected_bot/db"
	"go.uber.org/zap"
)

//var zLogger logger.Interface

var zLogger, _ = zap.NewDevelopment()

func TestDb(t *testing.T) {

	db := sdb.New(context.Background(), zLogger, "connected.db")
	err := db.InitDb()
	if err != nil {

		t.Fatalf("%v", err.Error())
	}

	testuser := sdb.User{
		TgID:        76774,
		Name:        "jon wick",
		IsTgPremium: false,
		IsInChannel: true,
		IsRemoved:   false,

		CalculatedQuota: 20,
	}
	uid, _ := uuid.NewV4()
	testconfig := sdb.Config{
		Id:        66,
		UUID:      uid,
		UserID:    247,
		InboundID: 3,
		Usage:     0,
		Active:    true,
		Quota:     20,
	}
	db.Create(&testconfig)
	db.Create(&testuser)
	var uppuser sdb.User
	db.First(&uppuser)
}
