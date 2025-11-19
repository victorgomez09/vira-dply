package assets

import "embed"

// Embed the entire dist directory from the web build
//
// If you have a web build output in the "dist" directory, uncomment the
// //go:embed line below and rebuild; otherwise FrontendFS will remain empty.
//
// //go:embed all:dist
var FrontendFS embed.FS
