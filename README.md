# Anisble MySQL_Query module

This extremely simple binary runs a sql query on a mysql db. Created to use inside AWX to avoid managing extra dependencies and creating a custom awx image.

## Usage

Download the appropriate binary from releases

Place in one of the following locations

- ~/.ansible/plugins/modules
- {{playbook_dir}}/library
- /usr/share/ansible/plugins/modules

## Arguments

`db_host` - Database Host
`port` - Database Port
`db_name` - Database to connect to
`username` - Username
`password` - Password
`query` - SQL Query string


## Example

```yaml
---
- hosts: localhost
  connection: local
  tasks:
    - name: MySQL Count Query
      mysql_query:
        db_host: 127.0.0.1
        port: 3306
        db_name: my-db
        username: mysql
        password: mysql
        query: "Select count(*) from test_table"
      register: count
    
    - debug:
        v: count.count

    - name: MySQL Query
      mysql_query:
        db_host: 127.0.0.1
        port: 3306
        db_name: my-db
        username: mysql
        password: mysql
        query: "Select * from test_table"
      register: query
    
    - debug:
        v: query.results
``` 

## Output

- `results` - List of Maps. Each row is converted to a map where column is the key and value is value

- `count` - Set when `count` is apart of the query
