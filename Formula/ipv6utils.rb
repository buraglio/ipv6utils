class Ipv6utils < Formula
  desc "A toolset for IPv6 utilities"
  homepage "https://github.com/buraglio/ipv6utils"
  url "https://github.com/buraglio/ipv6utils/archive/refs/tags/v4.tar.gz"
  sha256 "604c50eccc07db6c96b7efefbeb1e6b853e9681e0639297e119e3f34be367dd4"  
  depends_on "go"  #  Ensures Go is installed to build the project

  def install
    # Build the Go project
    system "go", "build", "-o", bin/"ipv6utils"

   
  end

  test do
    # A simple test to verify the install
    system "#{bin}/ipv6utils", "-h"
  end
end