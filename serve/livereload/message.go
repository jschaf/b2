package livereload

// helloRequest is the initial handshake message.
// {
//    command: 'hello',
//    protocols: [
//      'http://livereload.com/protocols/official-7',
//      'http://livereload.com/protocols/official-8',
//      'http://livereload.com/protocols/official-9',
//      'http://livereload.com/protocols/2.x-origin-version-negotiation',
//      'http://livereload.com/protocols/2.x-remote-control'
//    ],
//    serverName: 'LiveReload 2'
// }
type helloRequest struct {
	Command    string   `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}

func newHelloResponse() helloRequest {
	return helloRequest{
		Command:    "hello",
		Protocols:  []string{"http://livereload.com/protocols/official-7"},
		ServerName: "b2",
	}
}

func validateHelloRequest(req *helloRequest) bool {
	if req.Command != "hello" {
		return false
	}
	for _, clientP := range req.Protocols {
		for _, serverP := range newHelloResponse().Protocols {
			if clientP == serverP {
				return true
			}
		}
	}
	return false
}

// reloadResponse is a server-to-client message to reload a file.
//
// {
//    command: 'reload',
//    path: 'path/to/file.ext',
//    liveCSS: true
// }
type reloadResponse struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

func newReloadResponse(path string) reloadResponse {
	return reloadResponse{
		Command: "reload",
		Path:    path,
		LiveCSS: true,
	}
}

// alertResponse is a server-to-client message to display an alert on the
// client.
//
// {
//    command: 'alert',
//    message: 'HEY!'
//  }
type alertResponse struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

func newAlertResponse(m string) alertResponse {
	return alertResponse{
		Command: "alert",
		Message: m,
	}
}
