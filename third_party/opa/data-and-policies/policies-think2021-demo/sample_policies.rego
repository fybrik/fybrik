package dataapi.authz

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
	description := "If intent is Fraud Detection and role is Data Scientist and data residency is Turkey and processing location is not Turkey, redact columns with tag Confidential"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role == "Data Scientist"
	input.resource.metadata.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
	column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.Confidential]
	count(column_names) > 0
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description := "Deny because the role is not Data Scientist when intent is Fraud Detection"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role != "Data Scientist"
	input.resource.metadata.tags.residency == "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "If intent is Customer Behaviour Analysis and role is Business Analyst and data residency is Turkey, deny access to data if there is a column with tag Confidential"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role == "Business Analyst"
	input.resource.metadata.tags.residency == "Turkey"
	column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.Confidential]
	count(column_names) > 0
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "Deny because role is not Business Analyst when intent is Customer Behaviour Analysis"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role != "Business Analyst"
	input.resource.metadata.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "If data residency is Turkey but processing geography is not Turkey then deny writing"
	#user context and access type check
	input.action.actionType == "write"
	input.resource.metadata.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "If data residency is not Turkey and processing geography is neither Turkey nor EEA then deny writing"
	#user context and access type check
	input.action.actionType == "write"
	input.resource.metadata.tags.residency != "Turkey"
	input.action.processingLocation != "Turkey"
	input.action.processingLocation != "EEA"
}
