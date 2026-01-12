package wsdl

import (
    "fmt"
    "github.com/yhy0/ChYing/lib/gowsdl"
    "strings"
)

/**
   @author yhy
   @since 2024/12/28
   @desc //TODO
**/

func ParseWSDL(wsdl *gowsdl.WSDL) map[string][]string {
    soap := make(map[string][]string)
    // 先把参数名和类型存下来
    messages := GetMessage(wsdl)
    nameSpace := wsdl.TargetNamespace
    fmt.Println(wsdl.TargetNamespace)
    fmt.Println(messages)
    // 遍历服务和对应的方法
    for _, service := range wsdl.Service {
        for _, port := range service.Ports {
            // 对应 <service name="User_Service">
            // fmt.Printf("Port Name: %s, Binding: %s, Location: %s\n", port.Name, port.Binding, port.SOAPAddress.Location)
            //
            // 对应 <binding name="Username_Binding" type="tns:Username_PortType">
            for _, binding := range wsdl.Binding {
                // 对应 <portType name="Username_PortType">
                if port.Binding == "tns:"+binding.Name {
                    // 对应 <portType name="Username_PortType">
                    for _, portType := range wsdl.PortTypes {
                        // 根据  <operation name="Username"> 中的 input 去找参数名和类型
                        for _, operation := range portType.Operations {
                            // 获取参数   对应 <message name="UsernameRequest">
                            params := messages[strings.ReplaceAll(operation.Input.Message, "tns:", "")]
                            if operation.Input.SOAPBody.Namespace != "" {
                                nameSpace = operation.Input.SOAPBody.Namespace
                            }
                            
                            soapRequest := CreateSOAPRequest(nameSpace, operation.Name, params, "1.1")
                            // fmt.Println("SOAP Request:")
                            // fmt.Println(soapRequest)
                            soap[port.SOAPAddress.Location] = append(soap[port.SOAPAddress.Location], soapRequest)
                        }
                    }
                }
            }
        }
    }
    
    return soap
}
