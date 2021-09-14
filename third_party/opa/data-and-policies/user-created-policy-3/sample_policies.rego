package dataapi.authz

verdict[output] {
	count(rule) == 0
	output = {"action": {"name":"DenyAccess"}, "policy": "Deny by default"}
}
verdict[output] {
	count(rule) > 0
	output = rule[_]
}
rule[{"action": {"name":"RedactColumn", "columns": column_names}, "policy": description}] {
	description := "Columns with Confidential tag to be redacted before read action"
    #user context and access type check
	input.action.actionType == "read"
    input.context.intent == "Fraud Detection"
	input.context.role == "Data Scientist"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
	column_names := [input.resource.columns[i].name | input.resource.columns[i].tags.Confidential == "true"]
}
rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description := "Deny because the role is not Data Scientist when intent is Fraud Detection"
    #user context and access type check
    input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role != "Data Scientist"
	input.resource.tags.residency == "Turkey"
}

rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description = "Deny because columns have confidential tag"
    #user context and access type check
    input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role == "Business Analyst"
	input.resource.tags.residency == "Turkey"
    column_names := [input.resource.columns[i].name | input.resource.columns[i].tags.Confidential == "true"]
	count(column_names) > 0
}

rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description = "Deny because role is not Business Analyst when intent is Customer Behaviour Analysis"
    #user context and access type check
    input.action.actionType == "read"
	input.context.intent == "Customer Behaviour Analysis"
	input.context.role != "Business Analyst"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description = "Deny because role is not Data Scientist and intent is Fraud Detection but the processing geography is not Trukey"
    #user context and access type check
    input.action.actionType == "read"
	input.context.intent == "Fraud Detection"
	input.context.role != "Data Scientist"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description = "If data residency is Turkey but processing geography is not Turkey then deny writing"
    #user context and access type check
    input.action.actionType == "write"
	input.resource.tags.residency == "Turkey"
	input.action.processingLocation != "Turkey"
}

rule[{"action": {"name":"DenyAccess"}, "policy": description}] {
	description = "If data residency is not Turkey and processing geography is neither Turkey nor EEA then deny writing"
    #user context and access type check
	input.action.actionType == "write"
	input.resource.tags.residency != "Turkey"
	input.action.processingLocation != "Turkey"
	input.action.processingLocation != "EEA"
}
