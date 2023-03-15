package dataapi.authz

# This file contains rego rules that Fybrik uses for its policy decision.
#
# The following is explanation about rego rules to help understand the rules used in this file.
# For more information please refer to https://www.openpolicyagent.org/docs/latest/policy-language/
#
# - A rule in rego is conditional assignment such that the assignment takes place only if
#   all the conditions between the curly braces are true:
#   Assigment IF { CONDITIONS }
#
# - Rule order is irrelavant.
#
# - Multiple rules with the same name give logical or.
#
# - When OPA evaluates policies it binds data provided in the query to a global variable called input.
#
# - A policy desicison in rego is the value of variable. Fybrik asks for the value of `verdict` variable
#   by calling POST /v1/data/dataapi/authz/verdict { input }
#
# - Partial set rule assigns element to a set, for example:
#   deny["some string"] { # "some string" belongs to the set called `deny`
#      some_condition     # if some_condition is true
#   }
#
# - OPA policies for Fybrik have the following syntax:
#   `rule[{"action": <action>, "policy": <policy>}]` where `policy` is a string describing the action and `action` is JSON object
#   with the following form:
#   {
#        "name": <name>,
#        <property>: <value>,
#        <property>: <value>,
#        ...
#   }
#   * `name` is the name of the action. For example: "RedactAction"
#   * `property` is the name of the action property as defined in the [enforcement actions taxonomy](../concepts/taxonomy.md). For example: "columns".1
# This is actually partial set rule assignment of an object to a set called `rule`.

# If the conditions between the curly braces are true then Fybrik will get an object for the "Deny" action.
# This rule sets the deny action if variable named `rule` is empty.
verdict[output] {
	count(rule) == 0 # true if a variable named `rule` of type set is empty.
	output = {"action": {"name":"Deny", "Deny": {}}, "policy": "Deny by default"}
}

# `outputFormatted` object belongs to the set called verdict if all the conditions between the curly braces are true.
# In that case Fybrik will get all the actions that belong to set `rule`.
verdict[outputFormatted] {
	count(rule) > 0 # true if a variable named `rule` of type set is NOT empty of elements
	output = rule[_]
	actionName := output.action.name
	actionWithoutName := json.remove(output.action, ["name"])
	outputWithoutAction := json.remove(output, ["action"])
	actionFormatted := {"name": actionName, output.action.name: actionWithoutName}
	outputFormatted := object.union({"action": actionFormatted}, outputWithoutAction)
}

# If the conditions between the curly braces are true then assign an empty object to set `rule`.
# As the condition is false `rule` set is not assigned. This rule actually used to declare `rule` variable and
# if no other rule of name `rule` exists then the result is deny by default due to the `verdict` rule
# defined above.
rule[{}] { false }
