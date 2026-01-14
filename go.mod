module fatalder-termux

go 1.25.5

require (
	github.com/TriM-Organization/bedrock-world-operator v1.4.0
	github.com/Yeah114/WaterStructure v0.0.0-00010101000000-000000000000
	github.com/disintegration/imaging v1.6.2
	github.com/mholt/archiver/v3 v3.5.1
	github.com/Yeah114/blocks v0.0.0-00010101000000-000000000000
	golang.org/x/image v0.21.0
)

replace github.com/Yeah114/WaterStructure => ./modules/WaterStructure
replace github.com/Yeah114/blocks => ./modules/blocks
