package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	xerr "github.com/goclub/error"
	xhttp "github.com/goclub/http"
	xjson "github.com/goclub/json"
	sl "github.com/goclub/slice"
	sq "github.com/goclub/sql"
	xtime "github.com/goclub/time"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	xerr.PrintStack(run())
}

var db *sq.Database
var config = Config{}
var httpClient *xhttp.Client

func init() {
	yamlData, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(yamlData, &config); err != nil {
		panic(err)
	}
	if db, _, err = sq.Open("mysql", config.Mysql.FormatDSN()); err != nil {
		panic(err)
	}
	if err = db.Ping(context.TODO()); err != nil {
		panic(err)
	}
	httpClient = xhttp.NewClient(nil)
}

type ConfigKey struct {
	Key   string                  `yaml:"key"`
	Limit uint64                  `yaml:"limit"`
	API   map[string]ConfigKeyAPI `yaml:"api"`
}
type ConfigKeyAPI struct {
	Limit uint64 `yaml:"limit"`
}
type Config struct {
	Keys     []ConfigKey        `yaml:"keys"`
	Mysql    sq.MysqlDataSource `yaml:"mysql"`
	AuthKeys []string           `yaml:"auth_keys"`
}

type Proxy struct{}

func matchAPIPath(req *http.Request) (apiPath string) {
	path := "/" + req.URL.Path
	apiPath = path
	q := req.URL.Query()
	switch path {
	case "/ws/geocoder/v1/":
		switch true {
		case q.Get("location") != "":
			apiPath = path + "?location=*"
		case q.Get("address") != "":
			apiPath = path + "?address=*"
		}
	}
	return apiPath
}
func proxyRequest(resp http.ResponseWriter, req *http.Request) (canUseKey string, err error) {
	ctx := context.Background()
	apiPath := matchAPIPath(req)
	today := xtime.Today(xtime.LocChina)
	for _, v := range config.Keys {
		limit := v.API[apiPath].Limit
		if limit == 0 {
			limit = v.Limit
		}
		if err = db.InsertModel(ctx, &TlbsKeyUseRecord{
			Key:     v.Key,
			Date:    today,
			ApiPath: apiPath,
			Count:   0,
		}, sq.QB{
			UseInsertIgnoreInto: true,
		}); err != nil {
			return
		}
		var aff int64
		col := TlbsKeyUseRecord{}.Column()
		if aff, err = db.UpdateAffected(ctx, &TlbsKeyUseRecord{}, sq.QB{
			Where: sq.
				And(col.Key, sq.Equal(v.Key)).
				And(col.Date, sq.Equal(today)).
				And(col.ApiPath, sq.Equal(apiPath)).
				And(col.Count, sq.LT(limit)),
			Set:   sq.SetRaw(`count = count +1`),
			Limit: 1,
		}); err != nil {
			return
		}
		if aff == 1 {
			canUseKey = v.Key
			break
		}
	}
	if canUseKey == "" {
		err = xerr.Reject(1, "没有可用的key", true)
		return
	}
	return
}
func writeError(resp http.ResponseWriter, msg string) {
	var body []byte
	var err error
	if body, err = xjson.Marshal(map[string]any{
		"status":  1,
		"message": msg,
	}); err != nil {
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Write(body)
}
func (p Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if sl.Contains(config.AuthKeys, req.URL.Query().Get("key")) == false {
		writeError(resp, "key 错误")
		return
	}
	var canUseKey string
	var err error
	if canUseKey, err = proxyRequest(resp, req); err != nil {
		if reject, ok := xerr.AsReject(err); ok {
			writeError(resp, reject.Message)
			return
		}

		xerr.PrintStack(err)
		writeError(resp, "system error")
		return
	}
	newURL := &url.URL{
		Scheme:   "https",
		Host:     "apis.map.qq.com",
		Path:     req.URL.Path,     // 使用原始请求的 Path
		RawQuery: req.URL.RawQuery, // 使用原始请求的 Query 参数
	}
	q := newURL.Query()
	q.Set("key", canUseKey) // 修改 key 参数
	newURL.RawQuery = q.Encode()
	var r *http.Request
	if r, err = http.NewRequest(req.Method, newURL.String(), nil); err != nil {
		return
	}
	if proxyResp, bodyClose, statusCode, err := httpClient.Do(r); err != nil {
		writeError(resp, err.Error())
		return
	} else {
		defer bodyClose()
		resp.WriteHeader(statusCode)
		var b []byte
		if b, err = ioutil.ReadAll(proxyResp.Body); err != nil {
			return
		}
		resp.Write(b)
		return
	}

	return
}
func run() (err error) {
	r := xhttp.NewRouter(xhttp.RouterOption{})
	r.HandleFunc(xhttp.Route{xhttp.GET, "/favicon.ico"}, func(c *xhttp.Context) (err error) {
		return c.WriteBytes([]byte("/"))
	})
	r.PrefixHandler("/", &Proxy{})
	s := http.Server{
		Addr:    ":4324",
		Handler: r,
	}
	r.LogPatterns(&s)
	return s.ListenAndServe()
}
