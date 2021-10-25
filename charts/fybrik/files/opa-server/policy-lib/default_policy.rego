package dataapi.authz

verdict[output] {
	count(rule) == 0
	output = {"action": {"name":"Deny", "Deny": {}}, "policy": "Deny by default"}
}

verdict[outputFormatted] {
	count(rule) > 0
	output = rule[_]
	actionName := output.action.name
	actionWithoutName := json.remove(output.action, ["name"])
	outputWithoutAction := json.remove(output, ["action"])
	actionFormatted := {"name": actionName, output.action.name: actionWithoutName}
	outputFormatted := object.union({"action": actionFormatted}, outputWithoutAction)
}

rule[{}] { false }
