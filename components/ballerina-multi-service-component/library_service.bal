import ballerina/http;
import ballerina/log;

listener http:Listener libraryListener = new (9092);

service /wso2/services/library on libraryListener {

    resource function get books() returns string {
        log:printInfo("Service: wso2/services/library");
        return "Hello from wso2/services/library!";
    }
}
