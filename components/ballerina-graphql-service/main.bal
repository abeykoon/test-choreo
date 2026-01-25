import ballerina/graphql;

// Employee record type
type Employee record {|
    string id;
    string name;
    string department;
    int age;
|};

// Sample employee data
final Employee[] employees = [
    {id: "E001", name: "John Doe", department: "Engineering", age: 30},
    {id: "E002", name: "Jane Smith", department: "Marketing", age: 28},
    {id: "E003", name: "Bob Johnson", department: "Finance", age: 35}
];

service /graphql on new graphql:Listener(9090) {

    // Query to get an employee by ID
    resource function get getEmployee(string id) returns Employee? {
        foreach Employee emp in employees {
            if emp.id == id {
                return emp;
            }
        }
        return ();
    }
}
