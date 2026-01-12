package wsdl

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "github.com/yhy0/ChYing/lib/gowsdl"
    "net/http"
    "strings"
)

/**
   @author yhy
   @since 2024/12/19
   @desc //TODO
**/

type Parameter struct {
    Name string
    Type string
}

// CreateSOAPRequest 创建 SOAP 请求
func CreateSOAPRequest(xmlnsUrn, operationName string, params []Parameter, version string) string {
    bodyContent := make([]interface{}, 0)
    
    for _, value := range params {
        bodyContent = append(bodyContent, SOAPElement{Name: value.Name, Type: value.Type, Value: Payload(value.Type)})
    }
    
    soapenv := "http://schemas.xmlsoap.org/soap/envelope/"
    
    if version == "1.2" {
        soapenv = "http://www.w3.org/2003/05/soap-envelope"
    }
    envelope := Envelope{
        XmlnsXsi:     "http://www.w3.org/2001/XMLSchema-instance",
        XmlnsXsd:     "http://www.w3.org/2001/XMLSchema",
        XmlnsSoapenv: soapenv,
        Header:       Header{},
        Body: Body{
            Content: SOAPOperation{
                Name:    operationName,
                Content: bodyContent,
                // todo 这里需要根据不同的配置，设置不同的 encodingStyle
                EncodingStyle: "http://schemas.xmlsoap.org/soap/encoding/",
                Namespace:     "urn", // 这里设置命名空间
                Use:           "encoded",
            },
        },
    }
    if xmlnsUrn != "" {
        envelope.XmlnsUrn = xmlnsUrn
    }
    
    var buf bytes.Buffer
    enc := xml.NewEncoder(&buf)
    enc.Indent("  ", "    ")
    if err := enc.Encode(envelope); err != nil {
        fmt.Printf("error: %v\n", err)
    }
    return buf.String()
}

// CreateHTTPRequest 创建 HTTP 请求
func CreateHTTPRequest(endpoint, xmlnsUrn, operationName string, params []Parameter, version string) (*http.Request, error) {
    soapRequest := CreateSOAPRequest(xmlnsUrn, operationName, params, version)
    req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(soapRequest))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "text/xml; charset=utf-8")
    req.Header.Set("SOAPAction", fmt.Sprintf("%s#%s", xmlnsUrn, operationName))
    
    return req, nil
}

type Envelope struct {
    XMLName      xml.Name `xml:"soapenv:Envelope"`
    XmlnsXsi     string   `xml:"xmlns:xsi,attr"`
    XmlnsXsd     string   `xml:"xmlns:xsd,attr"`
    XmlnsSoapenv string   `xml:"xmlns:soapenv,attr"`
    XmlnsUrn     string   `xml:"xmlns:urn,attr,omitempty"`
    Header       Header   `xml:"soapenv:Header"`
    Body         Body     `xml:"soapenv:Body"`
}

type Header struct{}

type Body struct {
    Content SOAPOperation `xml:",any"`
}

type SOAPOperation struct {
    XMLName       xml.Name      `xml:""`
    Name          string        `xml:"-"`
    Content       []interface{} `xml:",any"`
    EncodingStyle string        // 新增字段，用于存储 encodingStyle
    Namespace     string        // 新增字段，用于存储 namespace
    Use           string        // 新增字段，用于存储 use
}

type SOAPElement struct {
    XMLName xml.Name `xml:"-"`
    Name    string   `xml:"-"`
    Type    string   `xml:"-"`
    Value   string   `xml:",chardata"`
}

func (e SOAPOperation) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
    if e.Namespace != "" {
        start.Name.Local = e.Namespace + ":" + e.Name // 设置命名空间
    } else {
        start.Name.Local = e.Name
    }
    
    // 根据 EncodingStyle、Use 动态添加属性
    if e.EncodingStyle != "" {
        start.Attr = append(start.Attr, xml.Attr{
            Name:  xml.Name{Local: "soapenv:encodingStyle"},
            Value: e.EncodingStyle,
        })
    }
    
    if err := enc.EncodeToken(start); err != nil {
        return err
    }
    for _, c := range e.Content {
        if err := enc.Encode(c); err != nil {
            return err
        }
    }
    return enc.EncodeToken(start.End())
}

func (e SOAPElement) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
    start.Name.Local = e.Name
    if e.Type != "" {
        start.Attr = append(start.Attr, xml.Attr{
            Name:  xml.Name{Local: "xsi:type"},
            Value: e.Type,
        })
    }
    
    if err := enc.EncodeToken(start); err != nil {
        return err
    }
    if err := enc.EncodeToken(xml.CharData(e.Value)); err != nil {
        return err
    }
    return enc.EncodeToken(start.End())
}

/*
GetMessage 把参数名和类型存下来
<message name="UsernameRequest">
<part name="username" type="xsd:string"/>
</message>
*/

func GetMessage(wsdl *gowsdl.WSDL) map[string][]Parameter {
    var messages = make(map[string][]Parameter)
    var elements = GetElements(wsdl)
    
    for _, message := range wsdl.Messages {
        for _, part := range message.Parts {
            if part.Element != "" {
                var element string
                if strings.Contains(part.Element, ":") {
                    element = strings.Split(part.Element, ":")[1]
                } else {
                    element = part.Element
                }
                
                messages[message.Name] = append(messages[message.Name], elements[element]...)
            } else {
                messages[message.Name] = append(messages[message.Name], Parameter{
                    Name: part.Name,
                    Type: part.Type,
                })
            }
        }
    }
    return messages
}

func GetElements(wsdl *gowsdl.WSDL) map[string][]Parameter {
    var elements = make(map[string][]Parameter)
    
    for _, schema := range wsdl.Types.Schemas {
        for _, element := range schema.Elements {
            // fmt.Printf("[%s] : %s\n", element.Name, element.Type)
            if element.ComplexType != nil {
                for _, sequence := range element.ComplexType.Sequence {
                    var parameter = Parameter{
                        Name: sequence.Name,
                    }
                    // fmt.Printf("  SubElement Name: %s\n", sequence.Name)
                    if sequence.SimpleType != nil {
                        parameter.Type = sequence.SimpleType.Restriction.Base
                    } else if sequence.Type != "" {
                        parameter.Type = sequence.Type
                    }
                    // fmt.Printf("[%s] : %s\n", element.Name, sequence.Type)
                    elements[element.Name] = append(elements[element.Name], parameter)
                }
            }
        }
    }
    return elements
}
