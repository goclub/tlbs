package tlbs_test

import (
	"context"
	"github.com/goclub/tlbs"
	"github.com/stretchr/testify/assert"
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
