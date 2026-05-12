import ballerinax/mysql;
import ballerina/sql;
import ballerinax/mysql.driver as _;

type DatabaseConfig record {|
    # If the MySQL server is secured, the username
    string user;
    # The password of the MySQL server for the provided username
    string password;
    # The name of the database
    string database;
    # Hostname of the MySQL server
    string host;
    # Port number of the MySQL server
    int port;
    # The `mysql:Options` configurations
    mysql:Options options?;
    # The `sql:ConnectionPool` configurations
    sql:ConnectionPool connectionPool?;
|};

# Database Client Configuration.
configurable DatabaseConfig dbConfig = ?;

# Database Client.
public final mysql:Client databaseClient = check new (...dbConfig);