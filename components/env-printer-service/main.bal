import ballerina/http;
import ballerina/log;
import ballerina/os;

// Prefixes of the env variables to print. CHOREO_ covers the default connection
// env names; MY_ covers names renamed through component.yaml configuration.env.
final string[] & readonly envPrefixes = ["CHOREO_", "MY_"];

// Schema-based configurable. It is required, so its value must come from the
// Config.toml that Choreo generates from the values entered in the console.
configurable string greeting = ?;

function init() {
    log:printInfo("Startup configurables", greeting = greeting);
    map<string> envs = matchingEnvs();
    if envs.length() == 0 {
        log:printWarn("No env variables found with the given prefixes", prefixes = envPrefixes);
    }
    foreach [string, string] [key, value] in envs.entries() {
        log:printInfo("Startup env", key = key, value = value);
    }
}

function matchingEnvs() returns map<string> {
    map<string> envs = {};
    foreach [string, string] [key, value] in os:listEnv().entries() {
        if envPrefixes.some(prefix => key.startsWith(prefix)) {
            envs[key] = value;
        }
    }
    return envs;
}

service / on new http:Listener(9090) {

    // Returns the configurable value and the matching env variables at request time.
    resource function get env() returns json {
        return {greeting, envs: matchingEnvs()};
    }
}
