// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package remote integrates the remote features of Viper.
package remote

import (
	"bytes"
	"io"
	"strings"

	"github.com/billyyoyo/viper"
)

type remoteConfigProvider struct{}

type KVPair struct {
	Key   string
	Value []byte
}

type KVPairs []*KVPair

type Response struct {
	Value []byte
	Error error
}

var InvokeConfigManager func(machines []string, username, password string) (RemoteConfigManager, error)

type RemoteConfigManager interface {
	Get(key string) ([]byte, error)
	List(key string) (KVPairs, error)
	Watch(key string, stop chan bool) <-chan *Response
}

func init() {
	viper.RemoteConfig = &remoteConfigProvider{}
}

func (rc remoteConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (rc remoteConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	resp, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(resp), nil
}

func (rc remoteConfigProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, nil
	}
	quit := make(chan bool)
	quitwc := make(chan bool)
	viperResponsCh := make(chan *viper.RemoteResponse)
	cryptoResponseCh := cm.Watch(rp.Path(), quit)
	// need this function to convert the Channel response form crypt.Response to viper.Response
	go func(cr <-chan *Response, vr chan<- *viper.RemoteResponse, quitwc <-chan bool, quit chan<- bool) {
		for {
			select {
			case <-quitwc:
				quit <- true
				return
			case resp := <-cr:
				vr <- &viper.RemoteResponse{
					Error: resp.Error,
					Value: resp.Value,
				}
			}
		}
	}(cryptoResponseCh, viperResponsCh, quitwc, quit)

	return viperResponsCh, quitwc
}

func getConfigManager(rp viper.RemoteProvider) (RemoteConfigManager, error) {
	var cm RemoteConfigManager
	var err error

	endpoints := strings.Split(rp.Endpoint(), ";")
	username, password := rp.AuthInfo()
	cm, err = InvokeConfigManager(endpoints, username, password)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
