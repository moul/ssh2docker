package ssh2docker

import (
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_CheckConfig(t *testing.T) {
	log.SetHandler(text.New(os.Stderr))

	Convey("Testing Server.CheckConfig", t, FailureContinues, func() {
		// FIXME: check with a script
		server, err := NewServer()
		So(err, ShouldBeNil)
		server.AllowedImages = []string{"alpine", "ubuntu:trusty", "abcde123"}

		So(server.CheckConfig(&ClientConfig{ImageName: "alpine"}), ShouldBeNil)
		So(server.CheckConfig(&ClientConfig{ImageName: "ubuntu:trusty"}), ShouldBeNil)
		So(server.CheckConfig(&ClientConfig{ImageName: "abcde123"}), ShouldBeNil)

		So(server.CheckConfig(&ClientConfig{ImageName: "abcde124"}), ShouldNotBeNil)
		So(server.CheckConfig(&ClientConfig{ImageName: "ubuntu:vivid"}), ShouldNotBeNil)
	})
}
