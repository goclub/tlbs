package tlbs

import (
	xerr "github.com/goclub/error"
	xjson "github.com/goclub/json"
	"log"
	"strings"
)

type District struct {
	Data            []DistrictItem
	LevelData       [3][]DistrictItem
	adcodeIndexHash map[string]int
	// nameIndexHash   map[string][]nameHash
}

//type nameHash struct {
//	Index int
//	Level Level
//}
type DistrictItem struct {
	ID string `json:"id"`
	// Deprecated: FullName 更全,部分区县 Name 是空的 FullName有名字
	Name     string               `json:"name"`
	Fullname string               `json:"fullname"`
	Pinyin   []string             `json:"pinyin"`
	Location DistrictItemLocation `json:"location"`
	Cidx     []int                `json:"cidx"`
}
type DistrictItemLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func NewDistrict(data []byte) (d District, err error) {
	jsondata := [3][]DistrictItem{}
	err = xjson.Unmarshal(data, &jsondata) // indivisible begin
	if err != nil {                        // indivisible end
		err = xerr.WrapPrefix("unmarshal district json error", err)
		return
	}
	d.LevelData = jsondata
	d.Data = append(d.Data, jsondata[0]...)
	d.Data = append(d.Data, jsondata[1]...)
	d.Data = append(d.Data, jsondata[2]...)
	d.adcodeIndexHash = map[string]int{}
	for i, data := range d.Data {
		d.adcodeIndexHash[data.ID] = i
	}
	//tempI := 0
	//d.nameIndexHash = map[string][]nameHash{}
	//indexHash := func(data []DistrictItem) {
	//	for i, item := range data {
	//		o := tempI + i
	//		d.nameIndexHash[item.Fullname] = append(d.nameIndexHash[item.Name], nameHash{
	//			Index: o,
	//			Level: LevelProvince,
	//		})
	//		tempI++
	//	}
	//}
	//indexHash(jsondata[0])
	//indexHash(jsondata[1])
	//indexHash(jsondata[2])
	//xjson.Print("indexHash", d.nameIndexHash)
	return
}
func (v District) FindByADCode(adcode string) (item DistrictItem, has bool) {
	index, hasIndex := v.adcodeIndexHash[adcode]
	if hasIndex == false {
		return
	}
	item = v.Data[index]
	has = true
	return
}
func (v District) ProvinceByADcode(adcode string) (province DistrictItem, has bool) {
	provinceADCode := adcode[0:2] + "0000"
	return v.FindByADCode(provinceADCode)
}
func (v District) CityByADcode(adcode string) (city DistrictItem, has bool) {
	cityADCode := adcode[0:4] + "00"
	return v.FindByADCode(cityADCode)
}
func (v District) Children(cidx []int) (children []DistrictItem) {
	if len(cidx) != 2 {
		return
	}
	start := cidx[0]
	end := cidx[1] + 1
	children = v.Data[start:end]
	return
}

type Level uint8

const (
	LevelProvince Level = 1
	LevelCity     Level = 2
	LevelDistrict Level = 3
	LevelOther    Level = 99 // adcode = 999999
)

func LevelSwitch() (province, city, district, other Level) {
	return LevelProvince, LevelCity, LevelDistrict, LevelOther
}

type Relationship struct {
	Fuzzy    bool   // 因为行政区域的更新不可预测,所以自带模糊查询. 如果 adcode=130526 找不到.将 Fuzzy 设为 true,然后按照 adcode=130500 查找,还是找不到按照 adcode=130000
	Adcode   string // Fuzzy = false 时 adcode 为入参 adcode, Fuzzy = true 会被修改
	Level    Level  // 省市区级别
	Province DistrictItem
	City     DistrictItem // 当 Level 为 LevelCity LevelDistrict 时不是 zero value
	District DistrictItem // 当 Level 为 LevelDistrict 时不是 zero value
}

// Relationship
// 根据 adcode 获取省市区信息
func (v District) Relationship(adcode string) (r Relationship, has bool) {
	if len(adcode) != 6 {
		return
	}
	if adcode == "999999" {
		r = Relationship{
			Fuzzy:    false,
			Adcode:   adcode,
			Level:    LevelOther,
			Province: DistrictItem{},
			City:     DistrictItem{},
			District: DistrictItem{},
		}
		has = true
		return
	}
	r, has = v.coreRelationship(adcode)
	fuzzy := false
	if has == false {
		fuzzy = true
		adcode = adcode[0:4] + "00"
		r, has = v.coreRelationship(adcode)
		if has == false {
			adcode = adcode[0:2] + "0000"
			r, has = v.coreRelationship(adcode)
		}
	}
	r.Fuzzy = fuzzy
	r.Adcode = adcode
	return
}
func (v District) coreRelationship(adcode string) (r Relationship, has bool) {
	itemIndex, hasItem := v.adcodeIndexHash[adcode]
	if hasItem == false {
		return
	}
	item := v.Data[itemIndex]
	switch {
	default:
		return
	case adcode[2:6] == "0000":
		r.Level = LevelProvince
		r.Province = item
	case adcode[4:6] == "00":
		r.Level = LevelCity
		r.City = item
		var hasProvince bool
		r.Province, hasProvince = v.ProvinceByADcode(adcode)
		if hasProvince == false {
			log.Print("goclub/tlbs: adcode(" + adcode + ") can not found province")
		}
	case adcode[4:6] != "00":
		r.Level = LevelDistrict
		r.District = item
		var hasProvince bool
		r.Province, hasProvince = v.ProvinceByADcode(adcode)
		if hasProvince == false {
			log.Print("goclub/tlbs: adcode(" + adcode + ")can not found province")
		}
		var hasCity bool
		r.City, hasCity = v.CityByADcode(adcode)
		if hasCity == false {
			// 直辖市
			r.City = r.Province
		}
	}
	has = true
	return
}

func (v District) RelationshipByAddress(addr string) (r Relationship, has bool) {
	type Info struct {
		Index int
		Item  DistrictItem
	}
	var p Info
	var c Info
	var district Info

	match := func(info *Info, remainAddr string, data []DistrictItem) {
		if remainAddr == "" {
			return
		}
		for _, item := range data {
			info.Index = strings.Index(remainAddr, item.Fullname)
			if info.Index >= 0 {
				info.Item = item
				break
			}
			if item.Name != "" {
				info.Index = strings.Index(remainAddr, item.Name)
				if info.Index >= 0 {
					info.Item = item
					break
				}
			}
		}
	}
	tempAddr := addr
	match(&p, addr, v.LevelData[0])
	if p.Item.ID != "" {
		parent := p
		tempAddr = strings.Replace(tempAddr, parent.Item.Fullname, "", 1)
		tempAddr = strings.Replace(tempAddr, parent.Item.Name, "", 1)
		match(&c, tempAddr, safeSlice(v.LevelData[1], parent.Item.Cidx))
	}
	if c.Item.ID != "" {
		parent := c
		tempAddr = strings.Replace(tempAddr, parent.Item.Fullname, "", 1)
		tempAddr = strings.Replace(tempAddr, parent.Item.Name, "", 1)
		match(&district, tempAddr, safeSlice(v.LevelData[2], parent.Item.Cidx))
	}
	// 直辖市的区
	if c.Item.ID != "" && strings.HasSuffix(c.Item.ID, "00") == false && district.Item.ID == "" {
		district = c
		c = p
	}
	if p.Item.ID != "" {
		has = true
		r.Province = p.Item
		r.Adcode = p.Item.ID
		r.Level = LevelProvince
	}
	if c.Item.ID != "" {
		r.City = c.Item
		r.Adcode = c.Item.ID
		r.Level = LevelCity
	}
	if district.Item.ID != "" {
		r.District = district.Item
		r.Adcode = district.Item.ID
		r.Level = LevelDistrict
	}

	return
}
