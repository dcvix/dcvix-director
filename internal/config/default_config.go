//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package config

import _ "embed"

//go:embed dcvix-director.conf.default
var defaultConfig []byte // Embed the file content as a byte slice
