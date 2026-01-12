package db

import (
    "github.com/yhy0/logging"
    "time"
)

/**
  @author: yhy
  @since: 2024/9/29
  @desc: //TODO
**/

type IPInfo struct {
    ID          int64  `gorm:"primary_key;auto_increment" json:"id"`
    Host        string `gorm:"index" json:"host"`
    Ip          string `gorm:"index" json:"ip"`
    AllRecords  string `gorm:"text" json:"all_records"`
    PortService string `gorm:"text" json:"port_service"`
    Type        string `json:"type"` // cdn 、waf、cloud
    Value       string `json:"value"`
    Cdn         bool   `json:"cdn"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func AddIPInfo(data *IPInfo) {
    if ExistIPInfo(data.Host) {
        UpdateIPInfo(data)
        return
    }
    globalDBTmp := GlobalDB.Model(&IPInfo{})
    err := globalDBTmp.Create(&data).Error
    
    if err != nil {
        logging.Logger.Errorln("AddSCopilot err:", err)
        return
    }
}

func GetAllIPInfo() []*IPInfo {
    var data []*IPInfo
    GlobalDB.Model(&IPInfo{}).Find(&data)
    return data
}

func ExistIPInfo(host string) bool {
    var data IPInfo
    GlobalDB.Model(&IPInfo{}).Where("host = ?", host).First(&data)
    if data.ID > 0 {
        return true
    }
    return false
}

func UpdateIPInfo(data *IPInfo) {
    err := GlobalDB.Model(&IPInfo{}).Where("host = ?", data.Host).Updates(data).Error
    if err != nil {
        logging.Logger.Errorln("UpdateIPInfo err:", err)
    }
}
