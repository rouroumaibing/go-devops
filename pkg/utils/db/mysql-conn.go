package db

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"k8s.io/klog/v2"
)

type MysqlConfig struct {
	DriverName string `json:"drivername"`
	User       string `json:"user"`
	Pwd        string `json:"pwd"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	DBName     string `json:"dbname"`
	MaxIdle    int    `json:"maxidle"`
	MaxConn    int    `json:"maxconn"`
}

func (m *MysqlConfig) Init() {
	err := orm.RegisterDriver(m.DriverName, orm.DRMySQL)
	if err != nil {
		klog.Error("注册数据库驱动失败：", m.DriverName, err)
		panic(err)
	}

	//default为数据库别名,DBName为实际的数据库名称
	//:MaxIdle 最大空闲数
	//:MaxConn 最大连接数
	//dbConn := "xxx:xxxxx@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4"
	//dbConn := "xxx:xxxxx@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&time_zone=%2B08%3A00" 时区设置为上海
	dbConn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4", m.User, m.Pwd, m.Host, m.Port, m.DBName)
	err = orm.RegisterDataBase("default", m.DriverName, dbConn, m.MaxIdle, m.MaxConn)
	if err != nil {
		klog.Error("数据库连接失败：", err)
		panic(err)
	}

	// 注册数据库表，修改为在models初始化时注册
	// orm.RegisterModel(new(User), new(Profile), new(Post))
	// orm.RegisterModel(new(User))

	//创建表
	//:数据库名称
	//:是否覆盖创建，true删除已有表创建，false只更新列的变更
	//:是否显示创建表的SQL语句
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		klog.Error("执行数据库同步失败：", err)
		panic(err)
	}

}
