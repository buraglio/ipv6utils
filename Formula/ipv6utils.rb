class Ipv6utils < Formula
  desc "A toolset for IPv6 utilities"
  homepage "https://github.com/buraglio/ipv6utils"
  url "https://github.com/buraglio/ipv6utils/archive/refs/tags/v3.tar.gz"
  sha256 "1b1d961f1861c269330468f758d8a0132fe5dc8ed3e37365db8d5a4ded432907"  
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