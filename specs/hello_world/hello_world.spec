# WEBSERVER SPEC FILE
# /////////////////////////////////////////////////

NAME = hello_world

VERSION = 1
REQUIRES = nginx, php

[PACKAGES]
	# NONE

[CONFIGS]
	debian_root = "/etc/"

[CONTENT]
	source = spec
	# can be extended to implement remote content fetch
	# example:
	# source = git
	# git_command = git clone ...
	debian_root = "/var/www/html/"

[COMMANDS]

