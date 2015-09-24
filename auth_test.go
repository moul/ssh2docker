package ssh2docker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_ImageIsAllowed(t *testing.T) {
	Convey("Testing Server.ImageIsAllowed", t, FailureContinues, func() {
		server, err := NewServer()
		So(err, ShouldBeNil)
		server.AllowedImages = []string{"alpine", "ubuntu:trusty", "abcde123"}

		So(server.ImageIsAllowed("alpine"), ShouldEqual, true)
		So(server.ImageIsAllowed("ubuntu:trusty"), ShouldEqual, true)
		So(server.ImageIsAllowed("abcde123"), ShouldEqual, true)

		So(server.ImageIsAllowed("abcde124"), ShouldEqual, false)
		So(server.ImageIsAllowed("ubuntu:vivid"), ShouldEqual, false)
	})
}
