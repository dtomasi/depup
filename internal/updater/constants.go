package updater

import "regexp"

var /* const */ VersionPattern = regexp.MustCompile(`(["']?)(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?)(["']?)`)
