package event

import (
	"fmt"

	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB is type of store
type DB struct {
	// c is the name of the database configuration
	c *options.DBConfig
	// Engine is the name of the database
	Engine *gorm.DB
}

const (
	DSNKey = "%v:%v@tcp(%v)/%v?charset=%v&parseTime=True&loc=Local"
)

// NewDB create new db
func NewDB(c *options.DBConfig) (*DB, error) {
	s := &DB{
		c: c,
	}
	engine, err := s.newEngine()
	if err != nil {
		return nil, err
	}
	s.Engine = engine
	return s, nil
}

// KeyDSN returns dsn
func KeyDSN(addr, userName, password, dataBase, charset string) string {
	return fmt.Sprintf(DSNKey, userName, password, addr, dataBase, charset)
}

// newEngine returns	mysql db
func (s *DB) newEngine() (*gorm.DB, error) {
	dsn := KeyDSN(s.c.Addr, s.c.UserName, s.c.Password,
		s.c.Database, s.c.Charset)
	log.Infof("dsn:%v", dsn)
	engine, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         uint(s.c.DefaultStringSize),
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger:                 GormLogger{},
		PrepareStmt:            s.c.EnablePrepareStmt,
		SkipDefaultTransaction: s.c.EnableSkipDefaultTransaction,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := engine.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(s.c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(s.c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(s.c.ConnMaxLifetime)

	return engine, nil
}
