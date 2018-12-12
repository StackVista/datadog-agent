#!/bin/bash

set -e

if [ -z ${STACKSTATE_AGENT_VERSION+x} ]; then
	# Pick the latest tag by default for our version.
	# STACKSTATE_AGENT_VERSION=$(inv version -u)
	STACKSTATE_AGENT_VERSION=$(cat $CI_PROJECT_DIR/version.txt)
	# But we will be building from the master branch in this case.
fi

echo $STACKSTATE_AGENT_VERSION

echo "$SIGNING_PUBLIC_KEY" | gpg --import
echo "$SIGNING_PRIVATE_KEY" > gpg_private.key
echo "$SIGNING_PRIVATE_PASSPHRASE" | gpg --batch --yes --passphrase-fd 0 --import gpg_private.key
echo "$SIGNING_KEY_ID"

ls $CI_PROJECT_DIR/outcomes/pkg/*.*

# Export your public key from your key ring to a text file.
gpg --export -a 'StackState' > RPM-GPG-KEY-stackstate

# Import your public key to your RPM DB
rpm --import RPM-GPG-KEY-stackstate

# Verify the list of gpg public keys in RPM DB
rpm -q gpg-pubkey --qf '%{name}-%{version}-%{release} --> %{summary}\n'

# Configure your ~/.rpmmacros file
# %_gpg_name  => Use the Real Name you used to create your key
echo "%_gpg_name StackState <info@stackstate.com>" > ~/.rpmmacros

# Sign your custom RPM package
chmod +x rpm-sign
./rpm-sign $SIGNING_PRIVATE_PASSPHRASE $CI_PROJECT_DIR/outcomes/pkg/*.rpm

# Check the signature to make sure it was signed
rpm --checksig $CI_PROJECT_DIR/outcomes/pkg/*.rpm
