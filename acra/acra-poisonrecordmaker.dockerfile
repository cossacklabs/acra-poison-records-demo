FROM golang:1.11

# Install dependencies
RUN apt-get update && apt-get -y install \
    libssl-dev

# Install libthemis, keep sources for later use
RUN ["/bin/bash", "-c", \
    "set -o pipefail && \
    curl -sSL https://pkgs.cossacklabs.com/scripts/libthemis_install.sh | \
        bash -s -- --yes --method source \
        --without-packing --without-clean"]

# Copy precompiled acra binaries and configs
COPY acra-poisonrecordmaker ./
ENTRYPOINT ["./acra-poisonrecordmaker"]
