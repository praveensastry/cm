module github.com/praveensastry/cm

go 1.15

require (
	github.com/DiSiqueira/GoTree v1.0.0
	github.com/hashicorp/hil v0.0.0-20200423225030-a18a1cd20038
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/sftp v1.12.0
	github.com/praveensastry/cm/terminal v0.0.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	gopkg.in/ini.v1 v1.61.0
)

replace github.com/praveensastry/cm/terminal => ./terminal
