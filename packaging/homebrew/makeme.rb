class Makeme < Formula
  desc "AI-powered 3D object generator CLI"
  homepage "https://github.com/ThomasVuNguyen/MakeMe"
  version "1.0.0"

  on_macos do
    on_arm do
      url "https://github.com/ThomasVuNguyen/MakeMe/releases/download/v#{version}/makeme-darwin-arm64.tar.gz"
      sha256 "0883c318c48333b4e3af13045edf6fe04d4a900ec115aac2027125f4a990ea06"
    end
  end

  def install
    libexec.install Dir["*"]

    %w[makeme stl2obj t3d].each do |exe|
      next unless (libexec/exe).exist?
      (bin/exe).write <<~EOS
        #!/bin/bash
        set -euo pipefail
        cd "#{libexec}"
        exec "#{libexec}/#{exe}" "$@"
      EOS
    end
  end

  def caveats
    <<~EOS
      Model assets are stored under "k/" relative to your working directory.
      Consider setting MAKEME_T3D if you install a different viewer.
    EOS
  end
end
