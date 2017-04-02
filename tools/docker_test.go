package tools

import (
	"testing"

	"github.com/ellcrys/util"
	docker "github.com/ncodes/go-dockerclient"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDocker(t *testing.T) {

	client, err := docker.NewClient(DockerEndpoint)
	if err != nil {
		t.Errorf("failed to create docker client. %s", err)
	}
	_ = client
	Convey("DockerTools", t, func() {

		Convey("DeleteContainer", func() {

			Convey("Should return error if container does not exists", func() {
				err := DeleteContainer(util.RandString(32), false, false, false)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrContainerNotFound)
			})

			Convey("Should successfully delete container", func() {

				container, err := client.CreateContainer(docker.CreateContainerOptions{
					Name: util.RandString(5),
					Config: &docker.Config{
						Cmd: []string{"ls"},
					},
				})
				So(err, ShouldBeNil)

				err = DeleteContainer(container.ID, false, false, false)
				So(err, ShouldBeNil)

				container, err = client.InspectContainer(container.ID)
				So(container, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "No such container")
			})

			Convey("Should successfully delete container and its image", func() {

				imageName := "busybox"
				container, err := client.CreateContainer(docker.CreateContainerOptions{
					Name: util.RandString(5),
					Config: &docker.Config{
						Cmd:   []string{"sleep "},
						Image: imageName,
					},
				})
				So(err, ShouldBeNil)

				err = DeleteContainer(container.ID, true, false, false)
				So(err, ShouldBeNil)

				container, err = client.InspectContainer(container.ID)
				So(container, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "No such container")

				img, err := client.InspectImage(imageName)
				So(img, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "no such image")
			})
		})
	})
}
