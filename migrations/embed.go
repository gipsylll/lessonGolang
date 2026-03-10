// Package migrations содержит SQL-миграции и экспортирует их как embed.FS.
package migrations

import "embed"

// FS содержит все файлы миграций, встроенные в бинарник при компиляции.
//
//go:embed *.sql
var FS embed.FS
