# tlbs

https://lbs.qq.com/ SDK 和本地增强函数

## NewDistrict | 行政区域
 
### Relationship | 上下级查询

根据 adcode 获取省市区信息

 ```go
package main

import "log"
import tlbs "github.com/goclub/tlbs"
func main() {
	log.Printf("%+v", run())
}
func run() (err error) {
	d, err := tlbs.NewDistrict(tlbs.DataDistrict20220707) // indivisible begin
	if err != nil { // indivisible end
	    return
	}
	r, has := d.Relationship("130102")
	log.Print(has)
	// true
	log.Print(r)
	/*
	{
	  "Fuzzy": false,
	  "Adcode": "130102",
	  "Level": 3,
	  "Province": {
	    "id": "130000",
	    "name": "河北",
	    "fullname": "河北省",
	    "pinyin": [
	      "he",
	      "bei"
	    ],
	    "location": {
	      "lat": 38.03599,
	      "lng": 114.46979
	    },
	    "cidx": [
	      32,
	      42
	    ]
	  },
	  "City": {
	    "id": "130100",
	    "name": "石家庄",
	    "fullname": "石家庄市",
	    "pinyin": [
	      "shi",
	      "jia",
	      "zhuang"
	    ],
	    "location": {
	      "lat": 38.04276,
	      "lng": 114.5143
	    },
	    "cidx": [
	      0,
	      21
	    ]
	  },
	  "District": {
	    "id": "130102",
	    "name": "",
	    "fullname": "长安区",
	    "pinyin": [],
	    "location": {
	      "lat": 38.03682,
	      "lng": 114.538955
	    },
	    "cidx": []
	  }
	}
	*/
	return
}
 ```


## JavaScript

**Vue组件**
https://github.com/2type/admin/blob/main/2type/module/region/index.js

**Tree 数据**
https://github.com/2type/admin/blob/main/2type/module/lbs/tree.js