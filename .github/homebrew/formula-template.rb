class Devplan < Formula
  desc "Development workflow automation tool"
  homepage "https://github.com/devplaninc/devplan-cli"
  version "${VERSION}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://${DO_SPACE_NAME}.${DO_SPACE_REGION}.digitaloceanspaces.com/releases/${GITHUB_REF_NAME}/devplan-darwin-arm64"
      sha256 "${SHA256_ARM64}"
    else
      url "https://${DO_SPACE_NAME}.${DO_SPACE_REGION}.digitaloceanspaces.com/releases/${GITHUB_REF_NAME}/devplan-darwin-amd64"
      sha256 "${SHA256_AMD64}"
    end
  end

  def install
    if Hardware::CPU.arm?
      bin.install "devplan-darwin-arm64" => "devplan"
    else
      bin.install "devplan-darwin-amd64" => "devplan"
    end
  end

  test do
    system "#{bin}/devplan", "--help"
  end
end
