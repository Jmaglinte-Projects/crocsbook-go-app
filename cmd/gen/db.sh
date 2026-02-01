#! /bin/bash -e 
 
OUTPUT_PATH=infra/mysql/lib

jet -source=mysql -dsn="root:@dm1n1234@tcp(localhost:3306)/db_crocs" -path=$OUTPUT_PATH
