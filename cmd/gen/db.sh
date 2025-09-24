#! /bin/bash -e 
 
OUTPUT_PATH=infra/mysql/lib

jet -source=mysql -dsn="root:super-secret-password@tcp(localhost:3369)/db_crocs" -path=$OUTPUT_PATH
