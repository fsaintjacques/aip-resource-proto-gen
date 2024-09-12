# aip-resource-proto-gen

This tiny helper will scaffold a protobuf service definition for a single
resource following the AIP guidances.

```
$ ./aip-resource-proto-gen --package acme.v1 --service=api.acme.com Organization
syntax = "proto3";

package acme.v1;

import "google/api/annotations.proto";
import "google/api/client.proto";
import "google/api/field_behavior.proto";
import "google/api/resource.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/field_mask.proto";
import "google/protobuf/timestamp.proto";

// Organization resource.
message Organization {
  option (google.api.resource) = {
    type: "api.acme.com/Organization",
    plural: "organizations",
    singular: "organization"
  };

  // The resource's name.
  string name = 1 [(google.api.field_behavior) = IDENTIFIER];

  // The resource's display name.
  string display_name = 2 [(google.api.field_behavior) = OPTIONAL];

  // The time at which the resource was created.
  google.protobuf.Timestamp create_time = 3 [(google.api.field_behavior) = OUTPUT_ONLY];

  // The time at which the resource was last updated.
  google.protobuf.Timestamp update_time = 4 [(google.api.field_behavior) = OUTPUT_ONLY];

  // Custom annotations defined by the caller.
  map<string, string> annotations = 5 [(google.api.field_behavior) = OPTIONAL];
}

// Request for GetOrganization method.
message GetOrganizationRequest {
  // The name of the resource to retrieve.
  string name = 1 [(google.api.field_behavior) = REQUIRED];
}

// Request for ListOrganization method.
message ListOrganizationRequest {
  // The maximum number of resources to return.
  int32 page_size = 1 [(google.api.field_behavior) = OPTIONAL];

  // The page token to use for pagination. Provide this to retrieve subsequent page
  string page_token = 2 [(google.api.field_behavior) = OPTIONAL];

  // The filter to apply to list results.
  string filter = 3 [(google.api.field_behavior) = OPTIONAL];

  // The order to list results by.
  string order_by = 4 [(google.api.field_behavior) = OPTIONAL];
}

// Response for ListOrganization method.
message ListOrganizationResponse {
  // The list of Organization resources.
  repeated Organization organizations = 1;

  // The token to retrieve the next page of results, or empty if there are no more results.
  string next_page_token = 2;
}

// Request for CreateOrganization method.
message CreateOrganizationRequest {
  // The ID to use for the resource. It will become the final component of the name.
  string organization_id = 1;

  // The Organization resource to create.
  Organization organization = 2 [(google.api.field_behavior) = REQUIRED];
}

// Request for UpdateOrganization method.
message UpdateOrganizationRequest {
  // The Organization resource to update. The resource must have
  Organization organization = 1 [(google.api.field_behavior) = REQUIRED];

  // The list of fields to update.
  google.protobuf.FieldMask update_mask = 2 [(google.api.field_behavior) = OPTIONAL];

  // If set to true, and the resource is not found, a new resource will be created.
  bool allow_missing = 3 [(google.api.field_behavior) = OPTIONAL];
}

// Request for DeleteOrganization method.
message DeleteOrganizationRequest {
  // The name of the resource to delete.
  string name = 1 [(google.api.field_behavior) = REQUIRED];

  // If set to true, and the resource is not found, no errors will be returned.
  bool allow_missing = 2 [(google.api.field_behavior) = OPTIONAL];
}

// Service for managing the Organization resource.
service OrganizationService {
  option (google.api.default_host) = "api.acme.com";

  // Get the Organization resource
  rpc GetOrganization(GetOrganizationRequest) returns (Organization) {
    option (google.api.http) = {get: "/v1/{name=organizations/*}"};

    option (google.api.method_signature) = "name";
  }

  // List the Organization resources
  rpc ListOrganization(ListOrganizationRequest) returns (ListOrganizationResponse) {
    option (google.api.http) = {get: "/v1/organizations"};

    option (google.api.method_signature) = "";
  }

  // Create a new Organization resource
  rpc CreateOrganization(CreateOrganizationRequest) returns (Organization) {
    option (google.api.http) = {
      post: "/v1/organizations",
      body: "organization"
    };

    option (google.api.method_signature) = "organization";
  }

  // Update the Organization resource
  rpc UpdateOrganization(UpdateOrganizationRequest) returns (Organization) {
    option (google.api.http) = {
      patch: "/v1/{organization.name=organizations/*}",
      body: "organization"
    };

    option (google.api.method_signature) = "organization,update_mask";
  }

  // Delete the Organization resource
  rpc DeleteOrganization(DeleteOrganizationRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {delete: "/v1/{name=organizations/*}"};

    option (google.api.method_signature) = "name";
  }
}
```