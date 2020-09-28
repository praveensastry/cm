# PHP SPEC FILE
# /////////////////////////////////////////////////

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
