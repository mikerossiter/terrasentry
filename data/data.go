// Package data ships the seed portability dataset embedded in the binary so a
// fresh install scores resources offline with no extra files. Users can still
// override it via portability.dataset_path in terrasentry.yaml.
package data

import _ "embed"

//go:embed portability.yaml
var Portability []byte
