# Waveform

Generate waveform data in JSON format from an audio file.
[![Build Status](https://travis-ci.org/delatech/waveform.svg?branch=master)](https://travis-ci.org/delatech/waveform)

The generated JSON file is optimized for space. Thus, the precision is
sufficient but not the most precise. The generated JSON structure only contains
integers and not floats. Before getting passed to
[Waveform.js](http://www.waveformjs.org/), values should be divided by `128`.

## Usage

`waveform /path/to/my/audio.mp3 > waveform.json`

## Requirements

This required [sox](http://sox.sourceforge.net/) and the `soxi` binary which
should be included with the `sox` distribution.

## Installation

### Debian

    echo "deb http://apt.delatech.net/ squeeze main" >> /etc/apt/sources.list
    curl https://raw.githubusercontent.com/delatech/gpg/master/delatech-public-key-sign.asc | apt-key add -
    apt-get update
    apt-get install sox libsox-fmt-mp3 waveform

### OSX

    brew tap delatech/delatech
    brew install sox
    brew install delatech/waveform

### Anywhere with a valid Go installation

    go get github.com/delatech/waveform

## LICENSE

Based on work from [Sunjin Lee](https://github.com/styner32/go-wave-to-json).
