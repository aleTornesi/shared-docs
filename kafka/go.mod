module kafka

go 1.25.0

require (
	github.com/lib/pq v1.11.2
	github.com/segmentio/kafka-go v0.4.47
)

require (
	github.com/aleTornesi/shared-docs/db v0.0.0
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)

replace github.com/aleTornesi/shared-docs/db v0.0.0 => ../db
