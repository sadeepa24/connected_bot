package db

import (
	"context"
	"io/fs"
	"os"
	"path"

	tgbotapi "github.com/sadeepa24/connected_bot/tg/tgbotapi"
	"go.uber.org/zap"

	sqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
	ctx      context.Context
	zLogger  *zap.Logger
	path     string
	Inilized bool
	dblogger logger.Interface
}

func New(ctx context.Context, logger *zap.Logger, path string) *Database {

	dbLogger := newdblgr(logger)

	db := Database{
		ctx:      ctx,
		zLogger:  logger,
		path:     path,
		Inilized: false,
		dblogger: dbLogger,
	}
	return &db
}

func (d *Database) InitDb() error {
	dir := path.Dir(d.path)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}
	c := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		FullSaveAssociations:   false,
	}
	d.DB, err = gorm.Open(sqlite.Open(d.path), c)
	
	
	//d.DB, err = gorm.Open(sqlite.Open(":memory:"), c)
	if err != nil {
		return err
	}
	if err = d.AutoMigrate(
		&User{},
		&Config{},
		&Inbound{},
		&Outbound{},
		&Admin{},
		&Adminchat{},
		&Metadata{},
		&UsageHistory{},
		&GiftLog{},
		&Reffral{},
		&Gift{},
		&Event{},
		&SboxConfigs{},
	); err != nil {
		return err
	}
	d.Inilized = true
	return nil
}

func (d *Database) Close() error {
	// TODO:
	// should check is all opration on db is over
	sqldb, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqldb.Close()

}

func (d *Database) AddUser(user *User) (*User, error) {
	return user, d.Model(&User{}).Create(user).Error

}

func (d *Database) GetUser(user *tgbotapi.User) (*User, error) {
	var getuser = &User{
		TgID: user.ID,
	}
	return getuser, d.Model(&User{}).First(getuser).Error
}

func (d *Database) DatabasePath() string {
	return d.path
}

