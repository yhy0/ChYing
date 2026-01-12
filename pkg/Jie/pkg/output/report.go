package output

import (
    _ "embed"
    "github.com/yhy0/logging"
    "os"
    "text/template"
)

/**
   @author yhy
   @since 2023/9/18
   @desc //TODO
**/

//go:embed vulnReport.html
var vulnReportTmpl []byte

var ReportMessageChan = make(chan VulMessage)

func GenerateVulnReport(filename string) {
    var vulMessages []VulMessage
    
    for vulMessage := range ReportMessageChan {
        // Append the received vulMessage to the list
        vulMessages = append(vulMessages, vulMessage)
        
        // Create the output file
        outputFile, err := os.Create(filename)
        if err != nil {
            logging.Logger.Errorln("Error creating output file:", err)
            return
        }
        
        // Execute the template and write the output to the file
        tmpl, err := template.New("vuln_report").Parse(string(vulnReportTmpl))
        if err != nil {
            logging.Logger.Errorln("Error parsing template:", err)
            return
        }
        
        err = tmpl.Execute(outputFile, struct {
            VulMessages []VulMessage
        }{
            VulMessages: vulMessages,
        })
        if err != nil {
            logging.Logger.Errorln("Error executing template:", err)
            return
        }
        
        outputFile.Close()
    }
}
