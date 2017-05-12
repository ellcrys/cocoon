package common

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"fmt"

	"os"

	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestFunc(t *testing.T) {
	Convey("Func", t, func() {
		Convey(".GetRPCErrDesc", func() {
			Convey("Should remove rpc error = 2 from error", func() {
				bs := GetRPCErrDesc(grpc.Errorf(1, "something bad happened"))
				So(string(bs), ShouldEqual, "something bad happened")
			})
		})

		Convey(".IsValidResName", func() {

			cases := [][]interface{}{
				[]interface{}{"lord.luggard", true},
				[]interface{}{"lord_luggard", true},
				[]interface{}{"lord-luggard", true},
				[]interface{}{"lordluggard", true},
				[]interface{}{"lord;luggard", false},
				[]interface{}{"lord@luggard", false},
			}

			for _, c := range cases {
				So(IsValidResName(c[0].(string)), ShouldEqual, c[1].(bool))
			}
		})

		Convey(".ReRunOnError", func() {
			Convey("Should successfully rerun function", func() {
				runCount := 0
				err := ReRunOnError(func() error {
					runCount++
					if runCount != 3 {
						return fmt.Errorf("error")
					}
					return nil
				}, 3, 0)
				So(err, ShouldBeNil)
				So(runCount, ShouldEqual, 3)
			})

			Convey("Expects rerun to fail after max re-run limit is reached without success", func() {
				runCount := 0
				err := ReRunOnError(func() error {
					runCount++
					return fmt.Errorf("error")
				}, 3, 0)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error")
				So(runCount, ShouldEqual, 3)
			})

			Convey("Should successfully rerun function with delay", func() {
				runCount := 0
				err := ReRunOnError(func() error {
					runCount++
					if runCount != 3 {
						return fmt.Errorf("error")
					}
					return nil
				}, 3, time.Millisecond*100)
				So(err, ShouldBeNil)
				So(runCount, ShouldEqual, 3)
			})
		})

		Convey(".CompareErr", func() {
			Convey("Should successfully match both errors", func() {
				So(CompareErr(errors.New("a"), errors.New("a")), ShouldEqual, 0)
			})
			Convey("Should return non zero (-1 or 1) if errors don't match", func() {
				So(CompareErr(errors.New("a"), errors.New("b")), ShouldEqual, -1)
				So(CompareErr(errors.New("b"), errors.New("a")), ShouldEqual, 1)
			})
		})

		Convey(".CapitalizeString", func() {
			Convey("Should successfully capitalize strings", func() {
				cases := [][]string{
					[]string{"the people are smiling", "The people are smiling"},
					[]string{"the people are smiling. they love the President", "The people are smiling. They love the President"},
				}
				for _, c := range cases {
					So(CapitalizeString(c[0]), ShouldEqual, c[1])
				}
			})
		})

		Convey(".ResolveFirewall", func() {
			Convey("Should return error if a rule's destination address could not be resolved", func() {
				_, err := ResolveFirewall([]types.FirewallRule{
					{Destination: "googleasadsa.com", DestinationPort: "80", Protocol: "udp"},
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "no such host")
			})

			Convey("Should successfully resolve rules destination", func() {
				rules, err := ResolveFirewall([]types.FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "facebook.com", DestinationPort: "80", Protocol: "tcp"},
				})
				So(err, ShouldBeNil)
				So(len(rules), ShouldBeGreaterThanOrEqualTo, 2)
				So(rules[0].Destination, ShouldNotEqual, "google.com")
				So(rules[1].Destination, ShouldNotEqual, "facebook.com")
			})

			Convey("Should handle nil rules by ignoring them", func() {
				rules, err := ResolveFirewall([]types.FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
				})
				So(err, ShouldBeNil)
				So(len(rules), ShouldBeGreaterThanOrEqualTo, 1)
				So(rules[0].Destination, ShouldNotEqual, "google.com")
			})
		})

		Convey(".RemoveASCIIColors", func() {
			Convey("Should successfully remove color", func() {
				So(RemoveASCIIColors([]byte("\033[1mHello Bold World!\033[0m")), ShouldResemble, []byte("Hello Bold World!"))
			})
		})

		Convey(".ReturnFirstIfDiffInt", func() {
			So(ReturnFirstIfDiffInt(1, 2), ShouldEqual, 1)
			So(ReturnFirstIfDiffInt(1, 1), ShouldEqual, 1)
			So(ReturnFirstIfDiffInt(0, 2), ShouldEqual, 2)
		})

		Convey(".ReturnFirstIfDiffString", func() {
			So(ReturnFirstIfDiffString("a", "xyz", false), ShouldEqual, "a")
			So(ReturnFirstIfDiffString("a", "a", false), ShouldEqual, "a")
			So(ReturnFirstIfDiffString("abc", "a", false), ShouldEqual, "abc")
			So(ReturnFirstIfDiffString("", "xyz", false), ShouldEqual, "")
			So(ReturnFirstIfDiffString("", "xyz", true), ShouldEqual, "xyz")
		})

		Convey(".GetAuthToken", func() {
			ctx := metadata.NewContext(context.Background(), nil)
			_, err := GetAuthToken(ctx, "bearer")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to get metadata from context")

			md := metadata.Pairs()
			ctx = metadata.NewIncomingContext(context.Background(), md)
			_, err = GetAuthToken(ctx, "bearer")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "authorization not included in context")

			md = metadata.Pairs("authorization", "abc")
			ctx = metadata.NewIncomingContext(context.Background(), md)
			_, err = GetAuthToken(ctx, "bearer")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "authorization format is invalid")

			md = metadata.Pairs("authorization", "basic abc")
			ctx = metadata.NewIncomingContext(context.Background(), md)
			_, err = GetAuthToken(ctx, "bearer")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Request unauthenticated with bearer")

			md = metadata.Pairs("authorization", "bearer abc")
			ctx = metadata.NewIncomingContext(context.Background(), md)
			token, err := GetAuthToken(ctx, "bearer")
			So(err, ShouldBeNil)
			So(token, ShouldEqual, "abc")
		})

		Convey(".HasEnv", func() {
			So(HasEnv("VAR_A", "VAR_B"), ShouldResemble, []string{"VAR_A", "VAR_B"})
			os.Setenv("VAR_A", "stuff")
			So(HasEnv("VAR_A", "VAR_B"), ShouldResemble, []string{"VAR_B"})
		})

		Convey(".MergeMapSlice", func() {
			s := []map[string]interface{}{
				{"name": "Ben", "age": 10, "skill": "running"},
				{"name": "Glen", "age": 15},
			}
			newMap := MergeMapSlice(s)
			So(newMap, ShouldHaveLength, 3)
			So(newMap["name"], ShouldEqual, "Glen")
			So(newMap["age"], ShouldEqual, 15)
			So(newMap["skill"], ShouldEqual, "running")
		})
	})
}
