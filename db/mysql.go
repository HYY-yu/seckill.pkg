package db

import (
	"context"
	"fmt"
	"time"

	"github.com/HYY-yu/werror"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _ Repo = (*dbRepo)(nil)

type Repo interface {
	i()
	GetDb(ctx context.Context) *gorm.DB
	DbClose() error
}

type dbRepo struct {
	Db *gorm.DB
}

type DBConfig struct {
	User            string
	Pass            string
	Addr            string
	Name            string
	MaxOpenConn     int
	MaxIdleConn     int
	ConnMaxLifeTime time.Duration

	ServerName string // 服务标识
}

func New(cfg *DBConfig) (Repo, error) {
	db, err := dbConnect(cfg)
	if err != nil {
		return nil, err
	}

	return &dbRepo{
		Db: db,
	}, nil
}

func (d *dbRepo) i() {}

func (d *dbRepo) GetDb(ctx context.Context) *gorm.DB {
	return d.Db.WithContext(ctx)
}

func (d *dbRepo) DbClose() error {
	sqlDB, err := d.Db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func dbConnect(cfg *DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		cfg.Pass,
		cfg.Addr,
		cfg.User,
		cfg.Name,
		true,
		"Local")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		return nil, werror.Wrap(err, fmt.Sprintf("[db connection failed] Database name: %s", cfg.Name))
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifeTime)

	// 使用插件
	err = db.Use(NewPlugin(cfg.ServerName, WithDBName(cfg.Name)))
	if err != nil {
		return nil, err
	}

	return db, nil
}
