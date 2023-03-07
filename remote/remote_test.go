package remote

import (
	"fmt"
	"github.com/spf13/viper"
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
	v.AddRemoteProvider("etcd3", "http://127.0.0.1:2379", "/configs/log.yml")
	v.AddRemoteProvider("etcd3", "http://127.0.0.1:2379", "/configs/upload.yml")
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
config:
remote:
enable: true
provider: etcd3
host: http://localhost:2379
keys:
	- /configs/upload.yml
	- /configs/log.yml
user: ${ETCD_USER:ro:ot}sdf
pwd: ${ETCD_PWD:ro:ot}sdf
refresh: true
`
	//reg := "$\\{{.}+\\}"
	reg := "\\$\\{(.)+\\}"
	exp, err := regexp.Compile(reg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	strs := exp.FindAllString(text, -1)
	for _, s := range strs {
		s = s[2 : len(s)-1]
		kv := strings.SplitN(s, ":", 2)
		fmt.Println(kv[0], kv[1])
	}
}
