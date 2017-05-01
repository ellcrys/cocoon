package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEnv(t *testing.T) {

	Convey("Env", t, func() {

		Convey(".GetFlags", func() {
			Convey("Should successfully get variable flags", func() {
				So(GetFlags("var@private"), ShouldResemble, []string{"private"})
				So(GetFlags("var@private,rand32"), ShouldResemble, []string{"private", "rand32"})
				So(GetFlags("var@"), ShouldResemble, []string{})
			})
		})

		Convey(".Process", func() {

			env := Env(map[string]string{"VAR_A": "1"})
			pub, priv := env.Process(false)
			So(len(priv), ShouldEqual, 0)
			So(len(pub), ShouldEqual, 1)
			So(pub["VAR_A"], ShouldEqual, "1")

			env = Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			pub, priv = env.Process(false)
			So(len(priv), ShouldEqual, 1)
			So(len(pub), ShouldEqual, 1)
			So(pub["VAR_A"], ShouldEqual, "1")
			So(priv["VAR_B"], ShouldEqual, "secret")

			env = Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			pub, priv = env.Process(true)
			So(len(priv), ShouldEqual, 1)
			So(len(pub), ShouldEqual, 1)
			So(pub["VAR_A"], ShouldEqual, "1")
			So(priv["VAR_B@private"], ShouldEqual, "secret")

			env = Env(map[string]string{"VAR_A": "1", "VAR_B@private,genRand16": "secret"})
			pub, priv = env.Process(false)
			So(len(priv), ShouldEqual, 1)
			So(len(pub), ShouldEqual, 1)
			So(pub["VAR_A"], ShouldEqual, "1")
			So(len(priv["VAR_B"]), ShouldEqual, 32)

			env = Env(map[string]string{"VAR_A@genRand32": "1", "VAR_B@private,genRand16": "secret"})
			pub, priv = env.Process(false)
			So(len(priv), ShouldEqual, 1)
			So(len(pub), ShouldEqual, 1)
			So(len(pub["VAR_A"]), ShouldEqual, 64)
			So(len(priv["VAR_B"]), ShouldEqual, 32)
		})

		Convey(".ProcessAsOne", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			all := env.ProcessAsOne(false)
			So(len(all), ShouldEqual, 2)
			So(all["VAR_A"], ShouldEqual, "1")
			So(all["VAR_B"], ShouldEqual, "secret")

			env = Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			all = env.ProcessAsOne(true)
			So(len(all), ShouldEqual, 2)
			So(all["VAR_A"], ShouldEqual, "1")
			So(all["VAR_B@private"], ShouldEqual, "secret")
		})
	})
}
