#!/bin/bash

url=${1:-http://www.unicode.org/Public/8.0.0/ucd/UnicodeData.txt}

pre="// autogenerated by genucd - do not edit

package ucd

// Source is the source of UCD data this package provides access to.
const Source = \`$url\`

// RawUCD provides access to the raw UCD data, with padding.
var RawUCD = rawUCD

var rawUCD = []byte(\`"

post='________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________________`)

const pad = 440  // must be at least twice as long as longest record.

var maxRaw = len(rawUCD) - pad
'

/bin/echo "$pre"
wget -O - $url
/bin/echo "$post"
