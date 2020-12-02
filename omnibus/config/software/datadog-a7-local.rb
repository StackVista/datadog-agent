
name "datadog-a7"
default_version "0.0.5"

dependency "pip"

build do
  ship_license "https://raw.githubusercontent.com/DataDog/datadog-checks-shared/master/LICENSE"
  pip "install --install-option=\"--install-scripts="\
      "#{windows_safe_path(install_dir)}/bin\" "\
      "wheel==0.30.0 "
  pip "install --install-option=\"--install-scripts="\
      "#{windows_safe_path(install_dir)}/bin\" "\
      "wheel==0.30.0 "\
      "#{name}==#{version} "\
      "configparser==3.5.0 " # configparser==3.5.0 pins a dependency of pylint->datadog-a7, later versions (up to v3.7.1) are broken.
  # TODO: all deps should be pinned.
end