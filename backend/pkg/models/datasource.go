package models

import (
	"errors"
	"fmt"
	"strings"

	"time"

	"github.com/datadefeat/datav/backend/pkg/db"
	"github.com/datadefeat/datav/backend/pkg/utils/simplejson"
)

var (
	ErrDataSourceNotFound                = errors.New("Data source not found")
	ErrDataSourceNameExists              = errors.New("Data source with the same name already exists")
	ErrDataSourceUidExists               = errors.New("Data source with the same uid already exists")
	ErrDataSourceUpdatingOldVersion      = errors.New("Trying to update old version of datasource")
	ErrDatasourceIsReadOnly              = errors.New("Data source is readonly. Can only be updated from configuration")
	ErrDataSourceAccessDenied            = errors.New("Data source access denied")
	ErrDataSourceFailedGenerateUniqueUid = errors.New("Failed to generate unique datasource id")
)

type DataSource struct {
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Version int    `json:"version"`

	Name        string `json:"name"`
	Type        string `json:"type"`
	TypeLogoUrl string `json:"typeLogoUrl"`
	Url         string `json:"url"`

	IsDefault bool `json:"isDefault"`

	User              string `json:"user"`
	Password          string `json:"password"`
	Database          string `json:"database"`
	BasicAuth         bool   `json:"basicAuth"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
	WithCredentials   bool   `json:"withCredentials"`

	JsonData         *simplejson.Json `json:"jsonData,omitempty"`
	SecureJsonData   *simplejson.Json `json:"secureJsonData,omitempty"`
	SecureJsonFields map[string]bool  `json:"secureJsonFields"`

	ReadOnly bool `json:"readOnly"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type DataSourceList []DataSource

func (slice DataSourceList) Len() int {
	return len(slice)
}

func (slice DataSourceList) Less(i, j int) bool {
	return strings.ToLower(slice[i].Name) < strings.ToLower(slice[j].Name)
}

func (slice DataSourceList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func QueryDataSource(id int64, name string) (*DataSource, error) {
	q := fmt.Sprintf("select id,name, uid, version, type, url, is_default, json_data,secure_json_data,basic_auth,created,updated from data_source where id='%d' or name='%s'", id, name)
	return queryDataSource(q)
}

func QueryDefaultDataSource() (*DataSource, error) {
	q := fmt.Sprintf("select id,name, uid, version, type, url, is_default, json_data,secure_json_data,basic_auth,created,updated from data_source where is_default='1'")
	return queryDataSource(q)
}

func queryDataSource(q string) (*DataSource, error) {
	var id int64
	var version int
	var uid, tp, url, name string
	var isDefault, basicAuth bool
	var created, updated time.Time
	var rawJSON []byte
	var rawSecureJson []byte

	err := db.SQL.QueryRow(q).Scan(&id, &name, &uid, &version, &tp, &url, &isDefault, &rawJSON, &rawSecureJson, &basicAuth, &created, &updated)
	if err != nil {
		return nil, err
	}

	jsonData := simplejson.New()
	err = jsonData.UnmarshalJSON(rawJSON)
	if err != nil {
		return nil, err
	}

	secureJsonData := simplejson.New()
	err = secureJsonData.UnmarshalJSON(rawSecureJson)
	if err != nil {
		return nil, err
	}

	sjMap, _ := secureJsonData.Map()
	secureJsonFields := make(map[string]bool)
	for k := range sjMap {
		secureJsonFields[k] = true
	}

	ds := &DataSource{
		Id:               id,
		Name:             name,
		Uid:              uid,
		Version:          version,
		Type:             tp,
		Url:              url,
		IsDefault:        isDefault,
		JsonData:         jsonData,
		SecureJsonData:   secureJsonData,
		BasicAuth:        basicAuth,
		Created:          created,
		Updated:          updated,
		SecureJsonFields: secureJsonFields,
	}

	return ds, nil
}
