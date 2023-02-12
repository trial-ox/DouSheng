package dao

import "log"

type TableUser struct {
	Id       int64
	Name     string
	Password string
}

func (tableUser TableUser) TableName() string {
	return "users"
}

//获取全部TableUser
func GetTableUserList() ([]TableUser, error) {
	tableUsers := []TableUser{}
	err := Db.Find(&tableUsers).Error
	if err != nil {
		log.Printf(err.Error())
		return tableUsers, err
	}
	return tableUsers, nil
}

//根据username获得TableUser
func GetTableUserByUsername(username string) (TableUser, error) {
	tableUser := TableUser{}
	err := Db.Where("name=?", username).First(&tableUser).Error
	if err != nil {
		log.Printf(err.Error())
		return tableUser, err
	}
	return tableUser, nil
}

func InsertTableUser(user *TableUser) bool {
	err := Db.Create(&user).Error
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

func GetTableUserById(id int64) (TableUser, error) {
	tableUser := TableUser{}
	err := Db.Where("id=?", id).First(&tableUser).Error
	if err != nil {
		log.Printf(err.Error())
		return tableUser, err
	}
	return tableUser, nil

}
