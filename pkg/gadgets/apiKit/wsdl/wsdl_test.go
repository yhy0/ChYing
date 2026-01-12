package wsdl

import (
    "encoding/xml"
    "fmt"
    "github.com/yhy0/ChYing/lib/gowsdl"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "testing"
)

/**
   @author yhy
   @since 2024/4/26
   @desc //TODO
**/

var data = []byte(`
<definitions name="UserService"
   targetNamespace="http://www.examples.com/wsdl/dvwsuserservice.wsdl"
   xmlns="http://schemas.xmlsoap.org/wsdl/"
   xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
   xmlns:tns="http://www.examples.com/wsdl/dvwsuserservice.wsdl"
   xmlns:xsd="http://www.w3.org/2001/XMLSchema">
 
   <message name="UsernameRequest">
      <part name="username" type="xsd:string"/>
      <part name="pwd" type="xsd:string"/>
   </message>

   <message name="UsernameResponse">
      <part name="username" type="xsd:string"/>
   </message>

   <portType name="Username_PortType">
      <operation name="Username">
         <input message="tns:UsernameRequest"/>
         <output message="tns:UsernameResponse"/>
      </operation>
   </portType>

   <binding name="Username_Binding" type="tns:Username_PortType">
      <soap:binding style="rpc"
         transport="http://schemas.xmlsoap.org/soap/http"/>
      <operation name="Username">
         <soap:operation soapAction="Username"/>
         <input>
            <soap:body
               encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"
               namespace="urn:examples:usernameservice"
               use="encoded"/>
         </input>

         <output>
            <soap:body
               encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"
               namespace="urn:examples:usernameservice"
               use="encoded"/>
         </output>
      </operation>
   </binding>

   <service name="User_Service">
      <documentation>WSDL File for DVWS User Service</documentation>
      <port binding="tns:Username_Binding" name="Username_Port">
         <soap:address
            location="http://dvws.local/dvwsuserservice/" />
      </port>
   </service>
</definitions>
`)

var data1 = []byte(`
<?xml version="1.0" encoding="utf-8"?>
<wsdl:definitions xmlns:s="http://www.w3.org/2001/XMLSchema"
                  xmlns:tns="http://www.mnb.hu/webservices/"
                  xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
                  xmlns:http="http://schemas.xmlsoap.org/wsdl/http/"
                  targetNamespace="http://www.mnb.hu/webservices/"
                  xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/">
  <wsdl:documentation xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/">MNB curreny exchange rate webservice.</wsdl:documentation>
  <wsdl:types>
    <s:schema elementFormDefault="qualified" targetNamespace="http://www.mnb.hu/webservices/">
      <s:element name="GetInfo">
        <s:complexType>
          <s:sequence>
            <s:element name="Id">
              <s:annotation>
                <s:documentation>comment</s:documentation>
              </s:annotation>
              <s:simpleType>
                <s:restriction base="s:string">
                  <s:minLength value="2"/>
                </s:restriction>
              </s:simpleType>
            </s:element>
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:element name="GetInfoResponse">
        <s:complexType>
          <s:sequence>
            <s:element minOccurs="0" maxOccurs="1" name="GetInfoResult" type="s:string">
                <s:annotation>
                    <s:documentation>this is a comment</s:documentation>
                </s:annotation>
            </s:element>
          </s:sequence>
        </s:complexType>
      </s:element>
      <s:complexType name="ResponseStatus">
        <s:sequence>
          <s:element name="status" minOccurs="0" maxOccurs="unbounded">
            <s:complexType>
              <s:simpleContent>
                <s:extension base="s:string">
                  <s:attribute name="code" use="required">
                    <s:simpleType>
                      <s:restriction base="s:string">
                        <s:enumeration value="UnrecognizedTrimName" />
                        <s:enumeration value="UnusedTrimName" />
                      </s:restriction>
                    </s:simpleType>
                  </s:attribute>
                </s:extension>
              </s:simpleContent>
            </s:complexType>
          </s:element>
        </s:sequence>
        <s:attribute ref="tns:responseCode"/>
      </s:complexType>
      <s:attribute name="responseCode">
        <s:simpleType>
          <s:restriction base="s:string">
            <s:enumeration value="Successful" />
            <s:enumeration value="Unsuccessful" />
            <s:enumeration value="ConditionallySuccessful" />
          </s:restriction>
        </s:simpleType>
      </s:attribute>
      <!-- element with local simple type -->
      <s:element name="elementWithLocalSimpleType">
        <s:annotation>
          <s:documentation>An element with a local simple type declaration including an enumeration.</s:documentation>
        </s:annotation>
        <s:simpleType>
          <s:restriction base="s:string">
            <s:enumeration value="enum1">
              <s:annotation>
                <s:documentation>First enum value</s:documentation>
              </s:annotation>
            </s:enumeration>
            <s:enumeration value="enum2">
              <s:annotation>
                <s:documentation>Second enum value</s:documentation>
              </s:annotation>
            </s:enumeration>
          </s:restriction>
        </s:simpleType>
      </s:element>
      <!-- element of type dateTime -->
      <s:element name="startDate" type="s:dateTime">
        <s:annotation>
          <s:documentation>The date and time when the process starts.</s:documentation>
        </s:annotation>
      </s:element>
    </s:schema>
  </wsdl:types>
  <wsdl:message name="GetInfoSoapIn">
    <wsdl:part name="parameters" element="tns:GetInfo" />
  </wsdl:message>
  <wsdl:message name="GetInfoSoapOut">
    <wsdl:part name="parameters" element="tns:GetInfoResponse" />
  </wsdl:message>
  <wsdl:portType name="MNBArfolyamServiceType">
    <wsdl:operation name="GetInfoSoap">
      <wsdl:input message="tns:GetInfoSoapIn"/>
      <wsdl:output message="tns:GetInfoSoapOut"/>
    </wsdl:operation>
  </wsdl:portType>
  <wsdl:binding name="MNBArfolyamBinding" type="tns:MNBArfolyamServiceType">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http" />
    <wsdl:operation name="GetInfoSoap">
      <wsdl:input>
        <soap:body use="literal" />
      </wsdl:input>
      <wsdl:output>
        <soap:body use="literal" />
      </wsdl:output>
    </wsdl:operation>
  </wsdl:binding>
  <wsdl:service name="MNBArfolyamService">
    <wsdl:documentation xmlns:wsdl="http://schemas.xmlsoap.org/wsdl/">MNB curreny exchange rate webservice.</wsdl:documentation>
    <wsdl:port name="MNBArfolyamServiceSoap" binding="tns:MNBArfolyamBinding">
      <soap:address location="http://example.org/" />
    </wsdl:port>
  </wsdl:service>
</wsdl:definitions>

`)

func TestUnmarshal(t *testing.T) {
    
    var wsdl gowsdl.WSDL
    err := xml.Unmarshal(data, &wsdl)
    if err != nil {
        t.Errorf("incorrect result\ngot:  %#v\nwant: %#v", err, nil)
    }
    
    fmt.Println("----")
    fmt.Println(ParseWSDL(&wsdl))
    var wsdl1 gowsdl.WSDL
    err = xml.Unmarshal(data1, &wsdl1)
    if err != nil {
        t.Errorf("incorrect result\ngot:  %#v\nwant: %#v", err, nil)
    }
    
    fmt.Println("----")
    fmt.Println(ParseWSDL(&wsdl1))
}

func TestUnmarshalHttp(t *testing.T) {
    conf.GlobalConfig.Http.MaxQps = 5
    response, err := httpx.Get("https://sosdirectws.sos.state.tx.us/ucc_ws/uccservice.asmx?wsdl")
    if err != nil {
        return
    }
    var wsdl gowsdl.WSDL
    err = xml.Unmarshal([]byte(response.Body), &wsdl)
    if err != nil {
        t.Errorf("incorrect result\ngot:  %#v\nwant: %#v", err, nil)
    }
    fmt.Println(ParseWSDL(&wsdl))
}

func TestDVWS(t *testing.T) {
    conf.GlobalConfig.Http.MaxQps = 5
    goWSDL, err := gowsdl.NewGoWSDL("http://127.0.0.1/dvwsuserservice?wsdl")
    if err != nil {
        return
    }
    
    fmt.Println(goWSDL.FindType("UsernameResponse"))
    
    fmt.Println(ParseWSDL(goWSDL.WSDL))
}
