package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type HordePlugin struct{}

func (p HordePlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "webmail_version=") || strings.Contains(value, "webmail4prod=") {
            return true
        }
    }

    if strings.Contains(body, "title=\"This site is powered by The Horde Application Framework.\" href=\"http://horde.org\">") {
        return true
    }

    if strings.Contains(body, "Powered by </font><a href=\"http://www.horde.org/\" TARGET=_blank>") {
        return true
    }

    if strings.Contains(body, "/themes/graphics/horde-power1.png\" alt=\"Powered by Horde\" title=\"\" />") {
        return true
    }

    if strings.Contains(body, "<html><body bgcolor=\"#aaaaaa\"><a href=\"icon_browser.php\">Application List</a><br /><br /><h2>Icons for My Account</h2>") {
        return true
    }

    if strings.Contains(body, "<script language=\"JavaScript\" type=\"text/javascript\" src=\"/hunter/js/enter_key_trap.js\"></script>") {
        return true
    }

    if strings.Contains(body, "<link href=\"/mail/mailbox.php?mailbox=INBOX\" rel=\"Top\" />") {
        return true
    }

    return false
}

func (p HordePlugin) Name() string {
    return "Horde - PHP Framework"
}
