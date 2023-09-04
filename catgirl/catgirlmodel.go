package catgirl

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"os"
)

type catgirldb gorm.DB

// 配置结构体
type catgirl struct {
	GroupId   int64  `gorm:"column:group_id" json:"group_id"`
	Name      string `gorm:"column:name"`
	Tall      int64  `gorm:"column:tall"`
	Weight    int64  `gorm:"column:weight"`
	Age       int64  `gorm:"column:age"`
	Character string `gorm:"column:character"`
}

func (catgirl) TableName() string {
	return "cat_girl"
}

// initializePush 初始化数据库
func initializeCat(dbpath string) *catgirldb {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	gdb.AutoMigrate(&catgirl{})
	return (*catgirldb)(gdb)
}

func (cdb *catgirldb) getAllGroupCatGirl(groupId int64) (catlist []catgirl) {
	db := (*gorm.DB)(cdb)
	db.Model(&catgirl{}).Find(&catlist, "group_id = ?", groupId)
	return
}

func (cdb *catgirldb) getCatGirlByName(groupId int64, name string) catgirl {
	db := (*gorm.DB)(cdb)
	var cat catgirl
	db.Model(&catgirl{}).Find(&cat, "name = ? and group_id = ?", name, groupId)
	return cat
}

func (cdb *catgirldb) insertOrUpdateCatGirl(bpMap map[string]any) (err error) {
	db := (*gorm.DB)(cdb)
	bp := catgirl{}
	data, err := json.Marshal(&bpMap)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &bp)
	if err != nil {
		return
	}
	if err = db.Model(&catgirl{}).First(&bp, "name = ? and group_id = ?", bp.Name, bp.GroupId).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Model(&catgirl{}).Create(&bp).Error
		}
	} else {
		err = db.Model(&catgirl{}).Where("name = ? and group_id = ?", bp.Name, bp.GroupId).Update(bpMap).Error
	}
	return
}
