# CLI for library [go-mysqldump](//github.com/Greg0/go-mysqldump)

Accepts arguments:

* `--connection` - Path to YAML file with connection configuration. Overwrites CLI connection arguments

Example:

```yaml
name: dump # Optional
address: 127.0.0.1:3306 # Optional 
username: root # Optional
password: root # Optional
dbname: database
```

* `--name` - Dump prefix. Default: `dump`
* `--addr` - Database address in format `host:port`
* `--user` - Database username 
* `--dbname` - Database name

Connection arguments `name`, `addr`, `user`, `pass`, `dbname` can be passed by file configuration in `connection` argument.

--- 

* `--ignore` - Path to file containing list of ignored tables. Table names should be placed in new lines. Table names are regex expressions.
* `--structOnly` - Path to file containing list of tables that will be dumped without data. Table names are regex expressions.

Example of file content:

```txt
_log$
_swap$
^secret
```

--- 

* `--output` - Path do direcotry where dump will be saved


## Examples

```sh
go-mysqldump-cli --connection conn.yml --ignore ignore.txt --structOnly structure.txt --output ./dumps 
```

```sh
go-mysqldump-cli --dbname database_to_dump --user root --pass root --addr 127.0.0.1:3306 --ignore ignore.txt --structOnly structure.txt --output ./dumps
```

