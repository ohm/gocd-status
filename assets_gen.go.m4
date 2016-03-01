// Copyright (c) 2016 Sebastian Ohm.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main
var (
define(`bt', changequote([,])[changequote([,])`changequote(`,')]changequote(`,'))
assetIndexHTML = []byte(bt()undivert(_assets/index.html)bt())
assetScriptJS = []byte(bt()undivert(_assets/script.js)bt())
)
