package db

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type dbLogger struct {
	zaplgr *zap.Logger
	//logger logger.Interface
}

func newdblgr(corelog *zap.Logger) *dbLogger {
	return &dbLogger{
		zaplgr: corelog,
	}
}

func (d *dbLogger) LogMode(lglevel logger.LogLevel) logger.Interface {
	return d
}

func (d *dbLogger) Info(ctx context.Context, msg string, val ...interface{}) {
	if ctx.Err() != nil {
		return
	}
	d.zaplgr.Info(msg, zap.Any("dbmsg", val))
}
func (d *dbLogger) Warn(ctx context.Context, msg string, val ...interface{}) {
	if ctx.Err() != nil {
		return
	}
	d.zaplgr.Warn(msg, zap.Any("dbmsg", val))
}
func (d *dbLogger) Error(ctx context.Context, msg string, val ...interface{}) {
	if ctx.Err() != nil {
		return
	}
	d.zaplgr.Error(msg, zap.Any("dbmsg", val))
}
func (d *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if ctx.Err() != nil {
		return
	}
	d.zaplgr.Info(begin.String())
	sql, row := fc()
	d.zaplgr.Info(sql, zap.Int("row", int(row)))
}
