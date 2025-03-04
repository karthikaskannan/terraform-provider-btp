---
page_title: "btp_directory_users Data Source - terraform-provider-btp"
subcategory: ""
description: |-
  Gets all users.
  Further documentation:
  https://help.sap.com/docs/btp/sap-business-technology-platform/user-and-member-management
---

# btp_directory_users (Data Source)

Gets all users.

__Further documentation:__
<https://help.sap.com/docs/btp/sap-business-technology-platform/user-and-member-management>

## Example Usage

```terraform
# look up all users which belong to the default identity provider on directory level
data "btp_directory_users" "defaultidp" {
  directory_id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
}

# look up all users which belong to a custom identity provider on directory level
data "btp_directory_users" "mycustomidp" {
  directory_id = "dd005d8b-1fee-4e6b-b6ff-cb9a197b7fe0"
  origin       = "my-custom-idp"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `directory_id` (String) The ID of the directory.

### Optional

- `origin` (String) The identity provider that hosts the user. Only needed for custom identity provider.

### Read-Only

- `id` (String, Deprecated) The ID of the directory.
- `values` (Set of String) The list of users assigned to the directory.