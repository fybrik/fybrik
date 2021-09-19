package dataapi.authz

verdict[output] {
	count(rule) == 0
	output = {"action": {"name":"Deny"}, "policy": "Deny by default"}
}

verdict[output] {
	count(rule) > 0
	output.action.name == "Deny"
	output = rule[_]
}

verdict[outputFormatted] {
	count(rule) > 0
	output.action.name != "Deny"
	output = rule[_]
	outputFormatted := {"action": {"name":output.action.name, output.action.name: output.action.columns}, "policy": output.policy}
}

rule[{"action": {"name":"RedactColumn", "columns": column_names}, "policy": description}] {
	description := "Columns with Confidential tag to be redacted before read action"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role == "Data Scientist"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
	column_names := [input.resource.columns[i].name | input.resource.columns[i].tags.Confidential]
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description := "Deny because the role is not Data Scientist when intent is Fraud Detection"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role != "Data Scientist"
	input.resource.tags.residency == "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "Deny because columns have confidential tag"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role == "Business Analyst"
	input.resource.tags.residency == "Turkey"
	column_names := [input.resource.columns[i].name | input.resource.columns[i].tags.Confidential]
	count(column_names) > 0
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "Deny because role is not Business Analyst when intent is Customer Behaviour Analysis"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role != "Business Analyst"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "Deny because role is not Data Scientist and intent is Fraud Detection but the processing geography is not Trukey"
	#user context and access type check
	input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role != "Data Scientist"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "If data residency is Turkey but processing geography is not Turkey then deny writing"
	#user context and access type check
	input.action.actionType == "write"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"Deny"}, "policy": description}] {
	description = "If data residency is not Turkey and processing geography is neither Turkey nor EEA then deny writing"
	#user context and access type check
	input.action.actionType == "write"
	input.resource.tags.residency != "Turkey"
	input.action.processingLocation != "Turkey"
	input.action.processingLocation != "EEA"
}