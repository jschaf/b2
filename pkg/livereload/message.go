package livereload

import (
	"bytes"
	"encoding/json"
)

type command string

const (
	helloCmd  command = "hello"
	reloadCmd command = "reload"
	alertCmd  command = "alert"
	infoCmd   command = "info"
)

type baseCmd struct {
	Command command `json:"command"`
}

// helloMsg is the initial handshake message.
//
//	{
//	   command: 'hello',
//	   protocols: [
//	     'http://livereload.com/protocols/official-7',
//	     'http://livereload.com/protocols/official-8',
//	     'http://livereload.com/protocols/official-9',
//	     'http://livereload.com/protocols/2.x-origin-version-negotiation',
//	     'http://livereload.com/protocols/2.x-remote-control'
//	   ],
//	   serverName: 'LiveReload 2'
//	}
type helloMsg struct {
	Command    command  `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}

func newHelloMsg() helloMsg {
	return helloMsg{
		Command:    "hello",
		Protocols:  []string{"http://livereload.com/protocols/official-7"},
		ServerName: "b2",
	}
}

func validateHelloMsg(req *helloMsg) bool {
	if req.Command != helloCmd {
		return false
	}
	for _, clientP := range req.Protocols {
		for _, serverP := range newHelloMsg().Protocols {
			if clientP == serverP {
				return true
			}
		}
	}
	return false
}

// reloadMsg is a server-to-client message to reload a file.
//
//	{
//	   command: 'reload',
//	   path: 'path/to/file.ext',
//	   liveCSS: true
//	}
type reloadMsg struct {
	Command command `json:"command"`
	Path    string  `json:"path"`
	LiveCSS bool    `json:"liveCSS"`
}

func newReloadMsg(path string) reloadMsg {
	return reloadMsg{
		Command: reloadCmd,
		Path:    path,
		LiveCSS: true,
	}
}

// alertMsg is a server-to-client message to display an alert on the
// client.
//
//	{
//	   command: 'alert',
//	   message: 'HEY!'
//	}
type alertMsg struct {
	Command command `json:"command"`
	Message string  `json:"message"`
}

func newAlertResponse(m string) alertMsg {
	return alertMsg{
		Command: alertCmd,
		Message: m,
	}
}

// infoMsg is a client-to-server message to share info about plugins and the
// current URL.
//
//	{
//	  command: 'info',
//	  plugins: {
//	    somePlugin: {
//	      version: '1.2'
//	      arbitraryKey: ['arbitrary value'],
//	    }
//	  },
//	  url: 'http://example.com',
//	}
type infoMsg struct {
	Command command                               `json:"command"`
	Plugins map[string]map[string]json.RawMessage `json:"plugins"`
	URL     string                                `json:"url"`
}

func formatInfoMsg(info *infoMsg) string {
	s := bytes.Buffer{}
	for name, value := range info.Plugins {
		s.WriteString(name + "{")
		for prop, v := range value {
			s.WriteString(" " + prop + ": ")
			jsonStr, _ := v.MarshalJSON()
			s.Write(jsonStr)
			s.WriteString(",")
		}
		s.WriteString(" } ")
	}
	return s.String()
}
