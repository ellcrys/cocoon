package launcher

import (
	"fmt"
)

// Launcher defines cocoon code
// deployment service.
type Launcher struct {
	failed    chan bool
	languages []Language
}

// NewLauncher creates a new launcher
func NewLauncher(failed chan bool) *Launcher {
	return &Launcher{
		failed: failed,
	}
}

// Launch starts a cocoon code
func (lc *Launcher) Launch(req *Request) {

}

// AddLanguage adds a new langauge to the launcher.
// Will return error if language is already added
func (lc *Launcher) AddLanguage(lang Language) error {
	if lc.GetLanguage(lang.GetName()) != nil {
		return fmt.Errorf("language already exist")
	}
	lc.languages = append(lc.languages, lang)
	return nil
}

// GetLanguage will return a langauges or nil if not found
func (lc *Launcher) GetLanguage(name string) Language {
	for _, l := range lc.languages {
		if l.GetName() == name {
			return l
		}
	}
	return nil
}

// GetLanguages returns all languages added to the launcher
func (lc *Launcher) GetLanguages() []Language {
	return lc.languages
}
