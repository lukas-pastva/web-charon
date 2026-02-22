package charon

import "embed"

//go:embed templates/public/*.html
var PublicTemplatesFS embed.FS

//go:embed templates/admin/*.html
var AdminTemplatesFS embed.FS

//go:embed static
var StaticFS embed.FS

//go:embed migrations/*.sql
var MigrationsFS embed.FS
