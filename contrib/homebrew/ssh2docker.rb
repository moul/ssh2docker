require "language/go"

class Ssh2docker < Formula
  desc "SSH server that creates a Docker container per connection (chroot++)"
  homepage "https://github.com/moul/ssh2docker"
  url "https://github.com/moul/ssh2docker/archive/v1.0.1.tar.gz"
  sha256 "8ff5355f7c0970433f7c5d1ff1a8ce9467e78c2011f4b886d4dca05d68b36786"

  head "https://github.com/moul/ssh2docker.git"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["CGO_ENABLED"] = "0"
    ENV.prepend_create_path "PATH", buildpath/"bin"

    mkdir_p buildpath/"src/github.com/moul"
    ln_s buildpath, buildpath/"src/github.com/moul/ssh2docker"
    Language::Go.stage_deps resources, buildpath/"src"

    # FIXME: update version
    system "go", "build", "-o", "ssh2docker", "./cmd/ssh2docker"
    bin.install "ssh2docker"

    # FIXME: add autocompletion
  end

  test do
    output = shell_output(bin/"ssh2docker --version")
    assert output.include? "ssh2docker version"
  end
end
