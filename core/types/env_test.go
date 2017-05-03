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

		Convey(".GetByFlag", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			So(len(env.GetByFlag("private")), ShouldEqual, 1)

			env = Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret", "VAR_C@private": "secret"})
			So(len(env.GetByFlag("private")), ShouldEqual, 2)

			env = Env(map[string]string{"VAR_A": "1"})
			So(len(env.GetByFlag("private")), ShouldEqual, 0)
		})

		Convey(".GetVarName", func() {
			So(GetVarName("MY_VAR@private"), ShouldEqual, "MY_VAR")
			So(GetVarName("MY_VAR"), ShouldEqual, "MY_VAR")
		})

		Convey(".Has", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			So(env.Has("VAR_A"), ShouldEqual, true)
			So(env.Has("VAR_B"), ShouldEqual, true)
			So(env.Has("VAR_C"), ShouldEqual, false)
		})

		Convey(".Set", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			env.Set("VAR_A", "2")
			So(env["VAR_A"], ShouldEqual, "2")

			env.Set("VAR_C", "3")
			So(env["VAR_C"], ShouldEqual, "3")
		})

		Convey(".Get", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})

			val, ok := env.Get("VAR_A")
			So(ok, ShouldEqual, true)
			So(val, ShouldEqual, "1")

			val, ok = env.Get("VAR_B")
			So(ok, ShouldEqual, true)
			So(val, ShouldEqual, "secret")

			val, ok = env.Get("VAR_C")
			So(ok, ShouldEqual, false)
			So(val, ShouldEqual, "")
		})

		Convey(".GetFull", func() {
			env := Env(map[string]string{"VAR_A": "1", "VAR_B@private": "secret"})
			k, v, ok := env.GetFull("VAR_B")
			So(k, ShouldEqual, "VAR_B@private")
			So(v, ShouldEqual, "secret")
			So(ok, ShouldEqual, true)

			k, v, ok = env.GetFull("VAR_B@private")
			So(k, ShouldEqual, "VAR_B@private")
			So(v, ShouldEqual, "secret")
			So(ok, ShouldEqual, true)
		})

		Convey(".ReplaceFlag", func() {
			newFlag, found := ReplaceFlag("VAR_A@private,public", "public", "protected")
			So(found, ShouldEqual, true)
			So(newFlag, ShouldEqual, "VAR_A@private,protected")

			newFlag, found = ReplaceFlag("VAR_A@private,public,abstract", "public", "protected")
			So(found, ShouldEqual, true)
			So(newFlag, ShouldEqual, "VAR_A@private,protected,abstract")

			newFlag, found = ReplaceFlag("VAR_A@private,public", "class", "abstract")
			So(found, ShouldEqual, false)
			So(newFlag, ShouldEqual, "VAR_A@private,public")

			newFlag, found = ReplaceFlag("VAR_A", "class", "abstract")
			So(found, ShouldEqual, false)
			So(newFlag, ShouldEqual, "VAR_A")
		})
	})
}
