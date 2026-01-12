package db

import (
    "github.com/yhy0/ChYing/pkg/decoder"
    "github.com/yhy0/logging"
    "time"
)

/**
  @author: yhy
  @since: 2024/9/10
  @desc: //TODO
**/

type Request struct {
    ID         uint `gorm:"primaryKey" gorm:"index" json:"id"`
    RequestId  uint `gorm:"index" json:"request_id"` // 请求 id
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Url        string `json:"url"`
    Host       string `gorm:"index" json:"host"` // 请求 host
    Path       string `gorm:"index" json:"path"` // 请求 uri
    RequestRaw string `json:"request_raw"`       // 完整请求
    RawMD5     string `json:"raw_md5"`
}

type Response struct {
    ID          uint `gorm:"primaryKey" json:"id"`
    RequestId   uint `gorm:"index" json:"request_id"` // 对应的请求 id，用于和响应关联
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Url         string `json:"url"`
    Path        string `json:"path"`
    Host        string `gorm:"index" json:"host"`
    ResponseRaw string `json:"response_raw"` // 完整请求响应
    ContentType string `json:"content_type"` // 响应内容类型
    RawMD5      string `json:"raw_md5"`
}

func AddRequest(req *Request, res *Response) {
    if ExistRequest(decoder.Md5(req.RequestRaw)) {
        return
    }

    err := RetryOnLocked("AddRequest", func() error {
        return GlobalDB.Model(&Request{}).Create(&req).Error
    }, 3)

    if err != nil {
        logging.Logger.Errorln("AddRequest err:", err)
        return
    }

    AddResponse(res)
}

func AddResponse(data *Response) {
    err := RetryOnLocked("AddResponse", func() error {
        return GlobalDB.Model(&Response{}).Create(&data).Error
    }, 3)

    if err != nil {
        logging.Logger.Errorln("AddResponse err:", err)
        return
    }
}

func ExistRequest(rawMD5 string) bool {
    var request Request
    globalDBTmp := GlobalDB.Model(&Request{})
    globalDBTmp.Where("raw_md5 = ?", rawMD5).First(&request)
    
    if request.ID == 0 {
        return false
    }
    
    return true
}

func GetResponse(pageNum int, pageSize int) (int64, []*Response) {
    var response []*Response
    var total int64
    GlobalDB.Model(&Response{}).Count(&total).Offset(pageNum).Limit(pageSize).Scan(&response)
    return total, response
}

func GetResponseByHost(host []string) []*Response {
    var response []*Response
    globalDBTmp := GlobalDB.Model(&Response{})
    for i, h := range host {
        if i > 0 {
            globalDBTmp = globalDBTmp.Or("host = ?", h)
        } else {
            globalDBTmp = globalDBTmp.Where("host = ?", h)
        }
    }
    globalDBTmp.Find(&response)
    
    return response
}

func GetTraffic(id int) (req *Request, res *Response) {
    GlobalDB.Model(&Request{}).Where("request_id = ?", id).First(&req)
    GlobalDB.Model(&Response{}).Where("request_id = ?", id).First(&res)
    return
}
