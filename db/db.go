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

//User changes

func (d *Database) AddUser(user *User) (*User, error) {
	return user, d.Model(&User{}).Create(user).Error

}

// Get userfrom db
func (d *Database) GetUser(user *tgbotapi.User) (*User, error) {
	var getuser = &User{
		TgID: user.ID,
	}
	//st := time.Now()
	err := d.Model(&User{}).First(getuser).Error
	//err := d.Model(getuser).Where(getuser).Error
	//d.zLogger.Info("ELpsed Time For GetUser Quary " + time.Since(st).String())
	return getuser, err
}

// return old user
func (d *Database) UpdateUser(newuser *tgbotapi.User, Id string) (*User, error) {
	return nil, nil
}

// return removed user
func (d *Database) RemoveUser(user *tgbotapi.User) (*User, error) {
	return nil, nil
}
func (d *Database) DatabasePath() string {
	return d.path
}

// Configs Handle

func (d *Database) Getadminchat() (map[int64]string, error) {
	var adminchats []Adminchat
	if err := d.Find(&adminchats).Error; err != nil {
		return map[int64]string{}, err
	}
	adchmap := make(map[int64]string, len(adminchats))
	for _, adminchat := range adminchats {
		adchmap[adminchat.Id] = adminchat.Name
	}

	return adchmap, nil
}

func (d *Database) QuaryAdmin(UserID int64) *Admin {
	var admin = &Admin{
		Id: UserID,
	}
	if err := d.First(admin); err != nil {
		return nil
	}
	return admin
}
