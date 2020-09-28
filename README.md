# cm
A minimalist configuration management tool.

## usage

To run this tool locally change into root of the project and run `GO111MODULE=on go run cm.go configure hello_world`

```bash
➜  cm git:(master) ✗ GO111MODULE=on go run cm.go
NAME:
   cm - Command Line Configuration Management System

USAGE:
   cm [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
   Praveen Sastry <sastry.praveen@gmail.com>

COMMANDS:
   list-hosts, lh     cm list-hosts
   configure, c       cm configure <spec/host name>
   add-host, ah       cm add-host
   delete-host, dh    cm delete-host
   list-specs, ls     cm list-specs
   describe-spec, ds  cm describe-spec
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## spec definition
A `.spec` file (short for specification), along with its `config` and `content` folders, contain the building blocks of a server configuration. Specs contain a list of packages to install, configuration and content files along with their destinations, and commands to run during the configuration job.

### spec format
```
NAME

VERSION
REQUIRES

[PACKAGES]

[CONFIGS]

[COMMANDS]
```

An example of a spec that installs php5:
```
NAME = php

VERSION = 1
REQUIRES =

[PACKAGES]
	apt_get = php5-fpm, php5-cli, php5-curl, php5-gd, php5-intl php5-mysql, php5-memcache, php5-mcrypt, php5-xmlrpc

[CONFIGS]
	debian_root = "/etc/"
	skip_interpolate = true

[COMMANDS]
	pre = "sudo apt-get install -y software-properties-common, sudo add-apt-repository -y ppa:ondrej/php, sudo apt-get update"
	post = "sudo service php5-fpm restart"

```

Specs can require other specs, to link smaller building blocks into more complex configurations.

### spec resolution

By default, **cm** will look for Specs in the following directories, in order, overwriting previously found specs with the same name:

1. ~/.cmspecs/
2. ./specs/

### inventory

By default, **cm** will look for hosts in `~/.cminventory` file. Inventory must stored in the below format
```
[host_alias]
        Host     = host_ip/host_domain_name
        Username = user_name
        Spec     = spec_name
        PassAuth = true/false 
```

Note: Only PassAuth is supported at the moment. SSHKeyAuth support is not implemented. So PassAuth must be true for now.