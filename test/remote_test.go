package test

import (
	"fmt"
	"github.com/billyyoyo/viper"
	//_ "github.com/billyyoyo/viper-plugins/etcd"
	"regexp"
	"strings"
	"testing"
	"time"
)

type LogConfig struct {
	File     LogFile `yaml:"file"`
	Level    string  `yaml:"level"`
	ErrStack bool    `yaml:"errStack"`
	Debug    bool    `yaml:"debug"`
}

func (w LogConfig) CanRefresh() bool {
	return true
}

func (w LogConfig) KeyName() string {
	return "log"
}

type LogFile struct {
	Path      string `yaml:"path"`
	Name      string `yaml:"name"`
	MaxSize   int    `yaml:"maxSize"`
	MaxBackup int    `yaml:"maxBackup"`
	MaxAge    int    `yaml:"maxAge"`
	Compress  bool   `yaml:"compress"`
}

type UploadConfig struct {
	Path    string `yaml:"path"`
	MaxSize int64  `yaml:"maxSize"`
	Timeout int64  `yaml:"timeout"`
}

func (w UploadConfig) CanRefresh() bool {
	return true
}

func (w UploadConfig) KeyName() string {
	return "upload"
}

func TestRemoteConfig(t *testing.T) {
	v := viper.New()
	v.AddRemoteProvider("etcd3", "http://127.0.0.1:2379", "/configs/log.yml", "", "")
	v.AddRemoteProvider("etcd3", "http://127.0.0.1:2379", "/configs/upload.yml", "", "")
	v.SetConfigType("yaml") // because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"
	err := v.ReadRemoteConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("start watch")
	go func() {
		logConf := LogConfig{}
		uploadConf := UploadConfig{}
		v.UnmarshalWithRefresh(&logConf)
		v.UnmarshalWithRefresh(&uploadConf)
		for {
			time.Sleep(3 * time.Second)
			fmt.Println("log.file.path:", logConf.File.Path,
				"upload.path:", uploadConf.Path)
		}
	}()
	v.WatchRemoteConfigOnChannel()
	running := make(chan int)
	<-running
}

func TestReg(t *testing.T) {
	text := `
dataSource:
  host: ${MYSQL_HOST:im.flyee.fun}
  port: ${MYSQL_PORT:3306}
  db: ${MYSQL_DB:jim}
  user: ${MYSQL_USER:root}
  password: ${MYSQL_PWD:hanjing}
  url: ${db.user}:${db.password}@tcp(${db.host}:${db.port})/${db.db}?charset=utf8&parseTime=True&loc=Local
  ddl: true
  debug: true
  slowSqlThreshold: 1000 # second
  tableNamePrefix: "im_"
  pool:
    maxOpenConns: 16
    maxIdleConns: 4
    connMaxLifeTime: 3600 #second
    connMaxIdleTime: 600 #second
`
	reg := "\\$\\{[0-9a-zA-Z_.:-]+\\}"
	exp, err := regexp.Compile(reg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	strs := exp.FindAllString(text, -1)
	for _, s := range strs {
		s = s[2 : len(s)-1]
		kv := strings.SplitN(s, ":", 2)
		fmt.Printf("%+v\n", kv)
	}
}
