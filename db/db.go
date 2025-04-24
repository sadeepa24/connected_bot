package db

import (
	"context"
	"errors"
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
	usageHistor *gorm.DB
	ctx      context.Context
	zLogger  *zap.Logger
	path     string
	historyPath string
	Inilized bool
	dblogger logger.Interface
}

func New(ctx context.Context, logger *zap.Logger, dbpath string, usagehistory string) (*Database, error) {

	dbLogger := newdblgr(logger)

	db := Database{
		ctx:      ctx,
		zLogger:  logger,
		path:     dbpath,
		Inilized: false,
		historyPath: usagehistory,
		dblogger: dbLogger,
	}
	return &db, nil
}

func (d *Database) InitDb() error {
	dir := path.Dir(d.path)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}
	dir = path.Dir(d.historyPath)
	err = os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}
	c := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		FullSaveAssociations:   false,
	}
	d.DB, err = gorm.Open(sqlite.Open(d.path), c)
	if err != nil {
		return err
	}
	//d.DB, err = gorm.Open(sqlite.Open(":memory:"), c)
	d.usageHistor, err = gorm.Open(sqlite.Open(d.historyPath), &gorm.Config{FullSaveAssociations: false})
	if err != nil {
		return err
	}
	if err :=  d.usageHistor.AutoMigrate(
		&UsageHistory{},
		&GiftLog{},
	); err != nil {
		return errors.New("usage history db init failed " + err.Error())
	}
	if err = d.AutoMigrate(
		&User{},
		&Config{},
		&Inbound{},
		&Outbound{},
		&Adminchat{},
		&Metadata{},
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
	usgsql, err := d.usageHistor.DB()
	if err == nil {
		usgsql.Close()
	}
	sqldb, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqldb.Close()
}

func (d *Database) AddUser(user *User) (*User, error) {
	return user, d.Model(&User{}).Create(user).Error
}

func (d *Database) CreateUsageHistory(user *UsageHistory) error {
	return d.usageHistor.Create(user).Error
}
func (d *Database) CreateGiftLog(gift *GiftLog) error {
	return d.usageHistor.Create(gift).Error
}
func (d *Database) CreateUsageHistories(users *[]UsageHistory) error {
	return d.usageHistor.Create(users).Error
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