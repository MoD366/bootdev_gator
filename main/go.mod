module github.com/MoD366/bootdev_gator

go 1.25.0

replace github.com/MoD366/bootdev_gator/internal/config v0.0.0 => ../internal/config
replace github.com/MoD366/bootdev_gator/internal/database v0.0.0 => ../internal/database

require github.com/MoD366/bootdev_gator/internal/config v0.0.0
require github.com/MoD366/bootdev_gator/internal/database v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.11.2 // indirect
)
