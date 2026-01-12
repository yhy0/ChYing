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

type SCopilot struct {
    ID        int64  `gorm:"primary_key;auto_increment" json:"id"`
    Host      string `gorm:"index" json:"host"`
    InfoCount int    `json:"info_count"`
    ApiCount  int    `json:"api_count"`
    VulnCount int    `json:"vuln_count"`
    JsonData  string `gorm:"type:text" json:"json_data"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func AddSCopilot(data *SCopilot) {
    if ExistSCopilot(data.Host) {
        UpdateSCopilot(data)
        return
    }
    globalDBTmp := GlobalDB.Model(&SCopilot{})
    err := globalDBTmp.Create(&data).Error
    
    if err != nil {
        logging.Logger.Errorln("AddSCopilot err:", err)
        return
    }
}

func GetSCopilotList() []*SCopilot {
    var data []*SCopilot
    GlobalDB.Model(&SCopilot{}).Find(&data)
    return data
}

func GetSCopilot(host string) []*SCopilot {
    globalDBTmp := GlobalDB.Model(&SCopilot{})
    var data []*SCopilot
    globalDBTmp.Where("host = ?", host).Find(&data)
    return data
}

func ExistSCopilot(host string) bool {
    var data SCopilot
    GlobalDB.Model(&SCopilot{}).Where("host = ?", host).First(&data)
    if data.ID > 0 {
        return true
    }
    return false
}

func UpdateSCopilot(data *SCopilot) {
    err := GlobalDB.Model(&SCopilot{}).Where("host = ?", data.Host).Updates(data).Error
    if err != nil {
        logging.Logger.Errorln("UpdateSCopilot err:", err)
    }
}
