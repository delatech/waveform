### WaveToJSON
Generate waveform data in json format from mp3 file
[![Build Status](https://travis-ci.org/styner32/go-wave-to-json.svg?branch=master)](https://travis-ci.org/styner32/go-wave-to-json)

### Installation

  go get github.com/styner32/go-wave-to-json

  This requires `sox`.

  install it via `brew` or `apt`
    $ brew install sox
  or
    $ sudo apt-get install sox libsox-fmt-mp3

### Example
  waveform.Generate("source.mp3", "result.json")
