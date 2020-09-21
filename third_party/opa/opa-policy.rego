
package extendedEnforcement

#possible results: allow, deny, transform

allow {
	count(deny) == 0
}

deny[used_policy] {
   not verify_access_type(input.type)
   used_policy := {
   		"used_policy" : construct_policy("unknown access type")
   }
} {
  
    not verify_purpose(input.purpose)
    used_policy := {
   		"used_policy" : construct_policy("unknown purpose")
   }
} {
    not verify_role(input.role)
    used_policy := {
   		"used_policy" : construct_policy("unknown role")
   }
} 
{
	input.type == "COPY"
    not verify_geography_for_copy(input.processing_geography)
    used_policy := {
   		"used_policy" : construct_policy("unknown geography to copy the data")
   }
}
	
transform[action] {
	description = "location data should be removed before copy"
	
    #deny has always higher priority than any other action
	allow
    
    #user context and access type check
    check_access_type(["COPY"])
    
    column_names := column_with_tag("location")
	action := {
    	"result":"Remove column",
        "args": {
        	"column name" : column_names[_]
        },
        "used_policy" : construct_policy(description)
    }
}  

transform[action] {
	description = "sensitive columns in health data should be removed"
	#deny has always higher priority than any other action
	allow
    
    #user context and access type check
    check_access_type(["COPY", "READ"])
    
    has_tag("HealthData")
    
    column_names := column_with_tag("SPI")
	action := {
    	"result":"Remove column",
        "args": {
        	"column name" : column_names[_]
        },
        "used_policy" : construct_policy(description)
    }
}         

transform[action] {
	description = "encrypt sensitive personal and health data on COPY out of united states for health data assets"
    
	#deny has always higher priority than any other action
	allow
    
    #user context and access type check
    check_access_type(["COPY"])
    
    has_tag("HealthData")
    input.destination != "United States" 
    
    column_names := column_with_any_tag(["SPI", "SMI"])
	action := {
    	"result":"Encrypt column",
        "args": {
        	"column name" : column_names[_]
        },
        "used_policy" : construct_policy(description)
    }
}  

#for dry run transactions.csv
transform[action] {
	description = "reduct columns with name nameOrig and nameDest  in datasets with Finance"
    
	#deny has always higher priority than any other action
	allow
    
    #user context and access type check
    check_access_type(["COPY", "READ"])
    
    has_tag("finance")
    
    column_names := column_with_any_name({"nameOrig", "nameDest"})
	action := {
    	"result":"Redact column",
        "args": {
        	"column name" : column_names[_]
        },
        "used_policy" : construct_policy(description)
    }
}  

# ################# Help functions ########################

allowed_access_types = ["READ", "COPY", "WRITE"]
allowed_purposes = ["analysis", "fraud-detection"]
allowed_roles = ["DataScientist", "Security"]

allowed_copy_destinations = ["NorthAmerica", "US"]

verify_access_type(access_type) {
		access_type == allowed_access_types[_]
}

verify_purpose(purpose) {
		purpose == allowed_purposes[_]
}

verify_role(role) {
	role == allowed_roles[_]
}

verify_geography_for_copy(destination) {
	destination == allowed_copy_destinations[_]  #here could be complex geo based computation
}

construct_policy(description) = policy {
	policy := {
    	"description":description
    }
}

has_tag(tag) {
	input.details.metadata.dataset_tags[_] == tag
}

column_with_tag(tag) = column_names {
	column_names := [column_name | input.details.metadata.components_metadata[column_name].tags[_] == tag]
}

column_with_any_tag(tags) = column_names {
	column_names := [column_name | input.details.metadata.components_metadata[column_name].tags[_] == tags[_]]
}

column_with_any_name(names) = column_names {
	all_column_names := {column_name | input.details.metadata.components_metadata[column_name] }
    column_names := all_column_names & names
}

check_access_type(access_types) {
	access_types[_] == input.type
}
