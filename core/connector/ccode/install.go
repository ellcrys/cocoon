package ccode

import logging "github.com/op/go-logging"

var log = logging.MustGetLogger("connect.installer")

// Install takes the cocoon code specified in
// the COCOON_CODE_URL environment variable,
// and installs it based on the language specified
// in COCOON_CODE_LANG
func Install() {
	log.Info("Ready to install cocoon code")
}
