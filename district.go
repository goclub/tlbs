package tlbs

import (
	xerr "github.com/goclub/error"
	xjson "github.com/goclub/json"
	"log"
)

type District struct {
	Data            []DistrictItem
	adcodeIndexHash map[string]int
}
type DistrictItem struct {
	ID       string               `json:"id"`
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
	d.Data = append(d.Data, jsondata[0]...)
	d.Data = append(d.Data, jsondata[1]...)
	d.Data = append(d.Data, jsondata[2]...)
	d.adcodeIndexHash = map[string]int{}
	for i, data := range d.Data {
		d.adcodeIndexHash[data.ID] = i
	}
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

// LevelSwitch
// example:
// province, city, district, other := tlbs.LevelSwitch()
//	switch v {
//	case province:
//		// TODO write code
//	case city:
//		// TODO write code
//	case district:
//		// TODO write code
//	case other:
//		// TODO write code
//	default:
//		err = xerr.New(fmt.Sprintf("tlbs.Level can not be %v", v))
//		return
//	}
func LevelSwitch() (province, city, district, other Level) {
	return LevelProvince, LevelCity, LevelDistrict, LevelOther
}

type Relationship struct {
	Fuzzy    bool   // ???????????????????????????????????????,????????????????????????. ?????? adcode=130526 ?????????.??? Fuzzy ?????? true,???????????? adcode=130500 ??????,????????????????????? adcode=130000
	Adcode   string // Fuzzy = false ??? adcode ????????? adcode, Fuzzy = true ????????????
	Level    Level  // ???????????????
	Province DistrictItem
	City     DistrictItem // ??? Level ??? LevelCity LevelDistrict ????????? zero value
	District DistrictItem // ??? Level ??? LevelDistrict ????????? zero value
}

// Relationship
// ?????? adcode ?????????????????????
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
			// ?????????
			r.City = r.Province
		}
	}
	has = true
	return
}
