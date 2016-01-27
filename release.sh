#!/bin/bash

function usage {
    echo -e "waveform release script\n"
    echo "Usage:"
    echo "  $0 version"
    exit 1
}

version=$1
if [ -z "$version" ]; then
    usage
fi

if  [ ! -d "bin" ]; then
    mkdir bin
fi



function xc {
    echo ">>> Cross compiling waveform"
    cd bin/
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o waveform-${version}-linux-amd64
    GOOS=linux GOARCH=386 go build -ldflags "-X main.version=${version}" -o waveform-${version}-linux-i386
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o waveform-${version}-darwin-amd64
    cd ../
}

function deb {
    arches="i386 amd64"
    for arch in $arches; do
        echo -e "\n>>> Creating debian package for ${arch}"
        fpm \
            -f \
            -s dir \
            -t deb \
            --vendor "DeLaTech" \
            --name   "waveform" \
            --description "An utility that creates JSON waveform from audiofile" \
            --version $version \
            -a $arch \
            -p ./bin/waveform-${version}-${arch}.deb \
            ./bin/waveform-${version}-linux-${arch}=/usr/bin/waveform
    done
}

function osx {
    echo -e "\n>>> Creating osx package"
    fpm \
        -f \
        -s dir \
        -t tar \
        --name   "waveform" \
        -p ./bin/waveform-darwin-${version}.tar \
        ./bin/waveform-${version}-darwin-amd64=/usr/bin/waveform
}

function publish_debian {
echo -e ">>> Publishing debian packages"
    aptly repo create delatech
    aptly repo add delatech bin/waveform-${version}-i386.deb
    aptly repo add delatech bin/waveform-${version}-amd64.deb
    aptly snapshot create delatech-waveform-${version} from repo delatech


    # for first tie use
    # aptly publish -distribution=squeeze snapshot delatech-waveform-${version} s3:apt.delatech.net:
    aptly publish switch squeeze s3:apt.delatech.net: delatech-waveform-${version}

}

function publish_homebrew {
    echo -e "\n>>> Publishing osx package"
    gzip -f ./bin/waveform-darwin-${version}.tar

    sha1sum=`sha1sum ./bin/waveform-darwin-${version}.tar.gz | awk '{print $1}'`
    aws s3 cp bin/waveform-darwin-${version}.tar.gz s3://release.delatech.net/waveform/waveform-${version}.tar.gz --acl=public-read

    cat <<EOF > $DELATECH_BREWTAP/Formula/waveform.rb
#encoding: utf-8

require 'formula'

class Waveform < Formula
    homepage 'https://github.com/delatech/waveform'
    version '${version}'

    url 'http://release.delatech.net.s3-website-eu-west-1.amazonaws.com/waveform/waveform-${version}.tar.gz'
    sha1 '${sha1sum}'

    depends_on :arch => :intel

    def install
        bin.install 'bin/waveform'
    end
end
EOF
    cd $DELATECH_BREWTAP
    git add Formula/waveform.rb
    git ci -m"Update waveform to v${version}"
    git push origin master
}

export AWS_ACCESS_KEY_ID=$AWS_DELATECH_S3_APT_KEY
export AWS_SECRET_ACCESS_KEY=$AWS_DELATECH_S3_APT_SECRET

xc
deb
publish_debian
osx
publish_homebrew
