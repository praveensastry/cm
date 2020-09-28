module github.com/praveensastry/cm

go 1.15

require (
	github.com/olekukonko/tablewriter v0.0.4
	github.com/praveensastry/cm/terminal v0.0.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/urfave/cli v1.22.4
	gopkg.in/ini.v1 v1.61.0
)

replace github.com/praveensastry/cm/terminal => ./terminal
