CLI_PREFIX = "/usr/local".freeze
CLI_URL = "https://api.github.com/repos/ajdnik/kube-cli/releases/latest".freeze

module Tty
  module_function

  def blue
    bold 34
  end

  def red
    bold 31
  end

  def reset
    escape 0
  end

  def bold(n = 39)
    escape "1;#{n}"
  end

  def underline
    escape "4;39"
  end

  def escape(n)
    "\033[#{n}m" if STDOUT.tty?
  end
end

class Array
  def shell_s
    cp = dup
    first = cp.shift
    cp.map { |arg| arg.gsub " ", "\\ " }.unshift(first).join(" ")
  end
end

def ohai(*args)
  puts "#{Tty.blue}==>#{Tty.bold} #{args.shell_s}#{Tty.reset}"
end

def warn(warning)
  puts "#{Tty.red}Warning#{Tty.reset}: #{warning.chomp}"
end

def system(*args)
  abort "Failed during: #{args.shell_s}" unless Kernel.system(*args)
end

def sudo(*args)
  args.unshift("-A") unless ENV["SUDO_ASKPASS"].nil?
  ohai "/usr/bin/sudo", *args
  system "/usr/bin/sudo", *args
end

def getc
  system "/bin/stty raw -echo"
  if STDIN.respond_to?(:getbyte)
    STDIN.getbyte
  else
    STDIN.getc
  end
ensure
  system "/bin/stty -raw echo"
end

def wait_for_user
  puts
  puts "Press RETURN to continue or any other key to abort"
  c = getc
  # we test for \r and \n because some stuff does \r instead
  abort unless (c == 13) || (c == 10)
end

class Version
  include Comparable
  attr_reader :parts

  def initialize(str)
    @parts = str.split(".").map { |p| p.to_i }
  end

  def <=>(other)
    parts <=> self.class.new(other).parts
  end
end

def force_curl?
  ARGV.include?("--force-curl")
end

def macos_version
  @macos_version ||= Version.new(`/usr/bin/sw_vers -productVersion`.chomp[/10\.\d+/])
end

def user_only_chmod?(d)
  return false unless File.directory?(d)
  mode = File.stat(d).mode & 0777
  # u = (mode >> 6) & 07
  # g = (mode >> 3) & 07
  # o = (mode >> 0) & 07
  mode != 0755
end

def chmod?(d)
  File.exist?(d) && !(File.readable?(d) && File.writable?(d) && File.executable?(d))
end

def chown?(d)
  !File.owned?(d)
end

def chgrp?(d)
  !File.grpowned?(d)
end

# Invalidate sudo timestamp before exiting (if it wasn't active before).
Kernel.system "/usr/bin/sudo -n -v 2>/dev/null"
at_exit { Kernel.system "/usr/bin/sudo", "-k" } unless $?.success?

# The block form of Dir.chdir fails later if Dir.CWD doesn't exist which I
# guess is fair enough. Also sudo prints a warning message for no good reason
Dir.chdir "/usr"

abort "Mac OS X too old." if macos_version < "10.10"
abort "Don't run this as root!" if Process.uid.zero?
abort <<-EOABORT unless `dsmemberutil checkmembership -U "#{ENV["USER"]}" -G admin`.include? "user is a member"
This script requires the user #{ENV["USER"]} to be an Administrator.
EOABORT

# Tests will fail if the prefix exists, but we don't have execution
# permissions. Abort in this case.
abort <<-EOABORT if File.directory?(CLI_PREFIX) && (!File.executable? CLI_PREFIX)
The install path, #{CLI_PREFIX}, exists but is not searchable. If this is
not intentional, please restore the default permissions and try running the
installer again:
    sudo chmod 775 #{CLI_PREFIX}
EOABORT

ohai "This script will install:"
puts "#{CLI_PREFIX}/bin/kube-cli"

group_chmods = %w[ bin bin/kube-cli].
               map { |d| File.join(CLI_PREFIX, d) }.
               select { |d| chmod?(d) }
chmods = group_chmods
chowns = chmods.select { |d| chown?(d) }
chgrps = chmods.select { |d| chgrp?(d) }
mkdirs = %w[bin].
         map { |d| File.join(CLI_PREFIX, d) }.
         reject { |d| File.directory?(d) }

unless group_chmods.empty?
  ohai "The following existing directories will be made group writable:"
  puts(*group_chmods)
end
unless chowns.empty?
  ohai "The following existing directories will have their owner set to #{Tty.underline}#{ENV["USER"]}#{Tty.reset}:"
  puts(*chowns)
end
unless chgrps.empty?
  ohai "The following existing directories will have their group set to #{Tty.underline}admin#{Tty.reset}:"
  puts(*chgrps)
end
unless mkdirs.empty?
  ohai "The following new directories will be created:"
  puts(*mkdirs)
end

wait_for_user if STDIN.tty? && !ENV["TRAVIS"]

if File.directory? CLI_PREFIX
  sudo "/bin/chmod", "u+rwx", *chmods unless chmods.empty?
  sudo "/bin/chmod", "g+rwx", *group_chmods unless group_chmods.empty?
  sudo "/usr/sbin/chown", ENV["USER"], *chowns unless chowns.empty?
  sudo "/usr/bin/chgrp", "admin", *chgrps unless chgrps.empty?
else
  sudo "/bin/mkdir", "-p", CLI_PREFIX
  sudo "/usr/sbin/chown", "root:wheel", CLI_PREFIX
end

unless mkdirs.empty?
  sudo "/bin/mkdir", "-p", *mkdirs
  sudo "/bin/chmod", "g+rwx", *mkdirs
  sudo "/bin/chmod", "755", *zsh_dirs
  sudo "/usr/sbin/chown", ENV["USER"], *mkdirs
  sudo "/usr/bin/chgrp", "admin", *mkdirs
end

ohai "Downloading and installing Kube CLI..."
Dir.chdir CLI_PREFIX do
  system "/bin/bash -o pipefail -c 'export TAG=$(/usr/bin/curl -fsSL #{CLI_URL} | /usr/bin/grep '\"tag_name\":' | /usr/bin/sed -E 's/.*\"([^\"]+)\".*/\1/') && /usr/bin/curl -fsSL --output bin/kube-cli.tar.gz https://github.com/ajdnik/kube-cli/releases/download/$TAG/kube-cli_darwin_amd64.tar.gz'"
  system "/bin/bash -o pipefail -c '/usr/bin/tar -xf bin/kube-cli.tar.gz && /bin/rm bin/kube-cli.tar.gz'" 
  system "/bin/chmod", "755", "bin/kube-cli"
end

warn "#{CLI_PREFIX}/bin is not in your PATH." unless ENV["PATH"].split(":").include? "#{CLI_PREFIX}/bin"

ohai "Installation successful!"
puts

# Use the shell's audible bell.
print "\a"
ohai "Next steps:"
puts "- Run `kube-cli --help` for more info"
