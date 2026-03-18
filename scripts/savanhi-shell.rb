# Homebrew Formula for Savanhi Shell
# Install with: brew install savanhi/tap/savanhi-shell

class SavanhiShell < Formula
  desc "Shell environment configuration tool with beautiful TUI"
  homepage "https://github.com/savanhi/shell"
  version "1.0.0"  # Update with each release
  license "MIT"

  # Specify the download URL and checksum
  # Update these with each release
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/savanhi/shell/releases/download/v1.0.0/savanhi-shell-1.0.0-darwin-amd64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_CHECKSUM"
  elsif OS.mac? && Hardware::CPU.arm?
    url "https://github.com/savanhi/shell/releases/download/v1.0.0/savanhi-shell-1.0.0-darwin-arm64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_CHECKSUM"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/savanhi/shell/releases/download/v1.0.0/savanhi-shell-1.0.0-linux-amd64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_CHECKSUM"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/savanhi/shell/releases/download/v1.0.0/savanhi-shell-1.0.0-linux-arm64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_CHECKSUM"
  end

  head do
    url "https://github.com/savanhi/shell.git", branch: "main"
    depends_on "go" => :build
  end

  # Dependencies
  depends_on "go" => :build if build.head?

  # Optional dependencies that enhance the experience
  depends_on "oh-my-posh" => :recommended
  depends_on "zoxide" => :recommended
  depends_on "fzf" => :recommended
  depends_on "bat" => :recommended
  depends_on "eza" => :recommended

  def install
    if build.head?
      # Build from source
      system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/savanhi-shell"
    else
      # Install pre-built binary
      bin.install "savanhi-shell"
    end

    # Install shell completions
    generate_completions_from_executable(bin/"savanhi-shell", "completion")

    # Install man pages (if available)
    man1.install "docs/savanhi-shell.1" if File.exist?("docs/savanhi-shell.1")
  end

  test do
    assert_match "Savanhi Shell", shell_output("#{bin}/savanhi-shell --version")
    
    # Test detection
    output = shell_output("#{bin}/savanhi-shell --detect")
    assert_match "OS:", output
    
    # Test help
    output = shell_output("#{bin}/savanhi-shell --help")
    assert_match "USAGE", output
  end

  def caveats
    <<~EOS
      Savanhi Shell has been installed!

      To complete the setup, you may want to install:
        brew install oh-my-posh zoxide fzf bat eza

      Run the interactive TUI:
        savanhi-shell

      For non-interactive installation:
        savanhi-shell --non-interactive --config config.json

      For more information:
        savanhi-shell --help

      Documentation: https://github.com/savanhi/shell#readme
    EOS
  end

  # Service for background operations (if needed in future)
  # service do
  #   run [opt_bin/"savanhi-shell", "daemon"]
  #   keep_alive true
  # end
end