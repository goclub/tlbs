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
	"strconv"
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
	keys := config.Keys
	sl.Shuffle(keys)
	for _, v := range keys {
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
	logValues := []any{canUseKey, req.URL.String(), newURL.String()}
	defer func() {
		if _, err = db.Exec(context.TODO(), "INSERT INTO tlbs_key_log (`key`, `r_url`, `t_url`, `body`) VALUES (?, ?, ?, ?)", logValues); err != nil {
			xerr.PrintStack(err)
			err = nil
			// 忽略插入错误
		}
	}()
	if proxyResp, bodyClose, statusCode, err := httpClient.Do(r); err != nil {
		logValues = append(logValues, err.Error())
		writeError(resp, err.Error())
		return
	} else {
		defer bodyClose()
		resp.WriteHeader(statusCode)
		var b []byte
		if b, err = ioutil.ReadAll(proxyResp.Body); err != nil {
			return
		}
		logValues = append(logValues, string(b))
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

// IDTlbsKeyUseRecord 用于类型约束
// 比如 userID managerID 都是 uint64,编码的时候如果传错就会出现bug
// 通过 IDTlbsKeyUseRecord 进行类型约束,如果参数不对编译器就会报错
type IDTlbsKeyUseRecord uint32

func NewIDTlbsKeyUseRecord(id uint32) IDTlbsKeyUseRecord {
	return IDTlbsKeyUseRecord(id)
}
func (id IDTlbsKeyUseRecord) Uint32() uint32 {
	return uint32(id)
}
func (id IDTlbsKeyUseRecord) IsZero() bool {
	return id == 0

}
func (id IDTlbsKeyUseRecord) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// 底层结构体,用于组合出 model
type TableTlbsKeyUseRecord struct {
	sq.WithoutSoftDelete
}

// TableName 给 TableName 加上指针 * 能避免 db.InsertModel(user) 这种错误， 应当使用 db.InsertModel(&user) 或
func (*TableTlbsKeyUseRecord) TableName() string { return "tlbs_key_use_record" }

// User model
type TlbsKeyUseRecord struct {
	Id      IDTlbsKeyUseRecord `db:"id" sq:"ignoreInsert"`
	Key     string             `db:"key"`
	Date    xtime.Date         `db:"date"`
	ApiPath string             `db:"api_path"`
	Count   uint32             `db:"count"`
	TableTlbsKeyUseRecord

	sq.DefaultLifeCycle
}

// AfterInsert 创建后自增字段赋值处理
func (v *TlbsKeyUseRecord) AfterInsert(result sq.Result) (err error) {
	var id uint64
	if id, err = result.LastInsertUint64Id(); err != nil {
		return
	}
	v.Id = IDTlbsKeyUseRecord(uint32(id))
	return
}

// Column dict
func (v TableTlbsKeyUseRecord) Column() (col struct {
	Id      sq.Column
	Key     sq.Column
	Date    sq.Column
	ApiPath sq.Column
	Count   sq.Column
}) {
	col.Id = "id"
	col.Key = "key"
	col.Date = "date"
	col.ApiPath = "api_path"
	col.Count = "count"

	return
}
