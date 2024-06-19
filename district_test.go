package tlbs_test

import (
	"context"
	xjson "github.com/goclub/json"
	"github.com/goclub/tlbs"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewDistrict(t *testing.T) {
	func() struct{} {
		// -------------
		var err error
		_ = err
		ctx := context.Background()
		_ = ctx
		d, err := tlbs.NewDistrict(tlbs.DataDistrict20220707) // indivisible begin
		assert.NoError(t, err)                                // indivisible end
		// 普通省市区:省
		{
			r, has := d.Relationship("130000")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "130000")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelProvince)
			assert.Equal(t, r.Province.ID, "130000")
			assert.Equal(t, r.City.ID, "")
			assert.Equal(t, r.District.ID, "")
		}
		// 普通省市区:市
		{
			r, has := d.Relationship("130100")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "130100")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelCity)
			assert.Equal(t, r.Province.ID, "130000")
			assert.Equal(t, r.City.ID, "130100")
			assert.Equal(t, r.District.ID, "")
		}
		// 普通省市区:区
		{
			r, has := d.Relationship("130102")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "130102")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelDistrict)
			assert.Equal(t, r.Province.ID, "130000")
			assert.Equal(t, r.City.ID, "130100")
			assert.Equal(t, r.District.ID, "130102")
		}
		// 特殊省市区:苏州工业园区
		{
			r, has := d.Relationship("320571")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "320571")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelDistrict)
			assert.Equal(t, r.Province.ID, "320000")
			assert.Equal(t, r.City.ID, "320500")
			assert.Equal(t, r.District.ID, "320571")
		}
		// 普通省市区:模糊搜索到市
		{
			r, has := d.Relationship("130198")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "130100")
			assert.Equal(t, r.Fuzzy, true)
			assert.Equal(t, r.Level, tlbs.LevelCity)
			assert.Equal(t, r.Province.ID, "130000")
			assert.Equal(t, r.City.ID, "130100")
			assert.Equal(t, r.District.ID, "")
		}
		// 普通省市区:模糊搜索到省
		{
			r, has := d.Relationship("139898")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "130000")
			assert.Equal(t, r.Fuzzy, true)
			assert.Equal(t, r.Level, tlbs.LevelProvince)
			assert.Equal(t, r.Province.ID, "130000")
			assert.Equal(t, r.City.ID, "")
			assert.Equal(t, r.District.ID, "")
		}
		// 直辖市:上海市
		{
			r, has := d.Relationship("310000")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "310000")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelProvince)
			assert.Equal(t, r.Province.ID, "310000")
			assert.Equal(t, r.City.ID, "")
			assert.Equal(t, r.District.ID, "")
		}
		// 直辖市:上海市-黄浦区
		{
			r, has := d.Relationship("310101")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "310101")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelDistrict)
			assert.Equal(t, r.Province.ID, "310000")
			assert.Equal(t, r.City.ID, "310000")
			assert.Equal(t, r.District.ID, "310101")
		}
		// 省市:广东省-东莞市
		{
			r, has := d.Relationship("441999")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "441999")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelDistrict)
			assert.Equal(t, r.Province.ID, "440000")
			assert.Equal(t, r.City.ID, "441900")
			assert.Equal(t, r.District.ID, "441999")
		}
		// 省直辖县:广东省-东莞市-东莞市
		{
			r, has := d.Relationship("441999")
			assert.Equal(t, has, true)
			assert.Equal(t, r.Adcode, "441999")
			assert.Equal(t, r.Fuzzy, false)
			assert.Equal(t, r.Level, tlbs.LevelDistrict)
			assert.Equal(t, r.Province.ID, "440000")
			assert.Equal(t, r.City.ID, "441900")
			assert.Equal(t, r.District.ID, "441999")
		}
		// -------------
		return struct{}{}
	}()
}

func TestRelationByAddress(t *testing.T) {
	func() struct{} {
		// -------------
		var err error
		_ = err
		ctx := context.Background()
		_ = ctx
		d, err := tlbs.NewDistrict(tlbs.DataDistrict20240319) // indivisible begin
		assert.NoError(t, err)                                // indivisible end
		type Case struct {
			Addr string
			Out  string
		}
		{
			r, h := d.RelationshipByAddress("新疆哈密市")
			assert.Equal(t, h, true)
			if b, err := xjson.Marshal(r); err != nil {
				panic(err)
			} else {
				assert.Equal(t, string(b), `{"Fuzzy":false,"Adcode":"650500","Level":2,"Province":{"id":"650000","name":"新疆","fullname":"新疆维吾尔自治区","pinyin":["xin","jiang"],"location":{"lat":43.793301,"lng":87.628579},"cidx":[430,454]},"City":{"id":"650500","name":"哈密","fullname":"哈密市","pinyin":["ha","mi"],"location":{"lat":42.819346,"lng":93.515053},"cidx":[2653,2655]},"District":{"id":"","name":"","fullname":"","pinyin":[],"location":{"lat":0,"lng":0},"cidx":[]}}`)
			}
		}
		list := []Case{
			{"新疆哈密市", "新疆维吾尔自治区,哈密市,."},
			{"北京", "北京市,,."},
			{"北京市", "北京市,,."},
			{"北京市海淀区", "北京市,北京市,海淀区."},
			{"上海市黄浦区", "上海市,上海市,黄浦区."},
			{"广东省", "广东省,,."},
			{"广东省广州市", "广东省,广州市,."},
			{"广东省广州市天河区", "广东省,广州市,天河区."},
			{"香港", "香港特别行政区,,."},
			{"澳门", "澳门特别行政区,,."},
			{"西藏", "西藏自治区,,."},
			{"宁夏", "宁夏回族自治区,,."},
			{"新疆", "新疆维吾尔自治区,,."},
			{"台湾", "台湾省,,."},
			{"台湾省", "台湾省,,."},
			{"台湾省台北市", "台湾省,台北市,."},
			{"内蒙古", "内蒙古自治区,,."},
			{"内蒙古自治区", "内蒙古自治区,,."},
			{"内蒙古赤峰市克什克腾旗", "内蒙古自治区,赤峰市,克什克腾旗."},
			{"山西省朔州市怀仁县", "山西省,朔州市,."},
			{"辽宁省沈阳市沈河区", "辽宁省,沈阳市,沈河区."},
		}
		for _, c := range list {
			r, has := d.RelationshipByAddress(c.Addr)
			outAddr := strings.Join([]string{r.Province.Fullname, r.City.Fullname, r.District.Fullname},
				",") + "."
			if has == false && c.Out != "" {
				assert.Equal(t, has, false, c.Addr)
				continue
			}
			if outAddr != c.Out {
				assert.Equal(t, c.Out, outAddr, c.Addr)
			}
		}
		return struct{}{}
	}()
}
