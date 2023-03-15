package dataapi.authz

# An empty object belongs to the set called `rule` if all the conditions between the curly braces are true.
# As the condition is true, `rule` set is always assigned and thus implies allow by default.
# defined above.
rule[{}] { true }
