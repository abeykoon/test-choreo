import ballerina/http;
import configFormIssue.database;

configurable string a = ?;
configurable string b = ?;


# A service representing a network-accessible API
# bound to port `9090`.
service / on new http:Listener(9090) {

    # A resource for generating greetings
    # + name - name as a string or nil
    # + return - string name with hello message or error
    resource function get greeting(string? name) returns string|error {


    _ = check database:databaseClient->execute(`
        INSERT INTO conference_qr
        (
            qr_id,
            info,
            description,
            coins,
            created_by
        )
        VALUES
        (
            a,
            b},
            v,
            s,
            s
        );
    `);
        // Send a response back to the caller.
        if name is () {
            return error("name should not be empty!");
        }
        return string `Hello, ${name}`;
    }
}
