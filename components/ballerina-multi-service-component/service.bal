import ballerina/http;
import ballerina/log;

listener http:Listener serviceListener = new (9091);

service /wso2/services on serviceListener {

    resource function get greeting() returns string {
        log:printInfo("Service: wso2/services");
        return "Hello from wso2/services!";
    }
}
