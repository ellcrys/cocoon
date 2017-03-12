package launcher

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MyLang struct {
}

func (ml *MyLang) GetName() string {
	return "ml"
}
func (ml *MyLang) GetImage() string {
	return ""
}
func (ml *MyLang) GetDownloadDestination() string {
	return ""
}
func (ml *MyLang) GetBuildScript() string {
	return ""
}

func (ml *MyLang) GetCopyDestination() string {
	return ""
}

func (ml *MyLang) GetSourceRootDir() string {
	return ""
}

func (ml *MyLang) GetRunScript() []string {
	return nil
}

func (ml *MyLang) RequiresBuild() bool {
	return false
}

func (ml *MyLang) SetBuildParams(map[string]interface{}) error {
	return nil
}

func TestLauncher(t *testing.T) {
	Convey("Launcher", t, func() {

		Convey("AddLanguage", func() {
			lc := NewLauncher(make(chan bool))
			Convey("should successfully add new language and return nil", func() {
				err := lc.AddLanguage(&MyLang{})
				So(err, ShouldBeNil)

				Convey("should return error if langauge has already been added", func() {
					err := lc.AddLanguage(new(MyLang))
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "language already exist")
				})
			})
		})

		Convey("GetLanguage", func() {
			lc := NewLauncher(make(chan bool))
			l := new(MyLang)
			err := lc.AddLanguage(l)
			So(err, ShouldBeNil)

			Convey("should return 1 language", func() {
				langs := lc.GetLanguages()
				So(len(langs), ShouldEqual, 1)
				So(langs[0], ShouldResemble, l)
			})
		})
	})
}
