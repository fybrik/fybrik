import re

def define_env(env):
  "Hook function"

  @env.macro
  def fybrik_version(version):
    if version == "__Release__":
        return "master"
    return version

  @env.macro
  def fybrik_image_version(version):
    if version == "__Release__":
        return "master"
    return version[1:]

  @env.macro
  def fybrik_version_flag(version):
    if re.match('^v[0-9]+\.[0-9]+(\.[0-9]+)*', version):
      return  '--version ' + version[1:]
    return "--version master"

  @env.macro
  def arrow_flight_module_version(version, arrow_flight_version):
    if version in arrow_flight_version:
        return arrow_flight_version[version]
    major_version = version[:4]
    if major_version in arrow_flight_version:
        return arrow_flight_version[major_version]
    return  "latest"
