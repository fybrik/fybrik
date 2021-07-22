package data_policies

#this file assumes input to be provided in specific format, in this case how data mesh provides it
#similar file can be built for Egeria, at least for the metadata part, or any other catalog when we show how the input should be  parsed correctly

#Example structure:
# {
# 	"name": "<name>"
# 	"destination": "<destination>",
# 	"processing_geography": "<processing_geography>",
# 	"purpose": "<purpose>",
# 	"role": "<role>",
# 	"type": "<access type>",
# 	"details": {
# 		"data_format": "<data_format>",
# 		"data_store": {
# 			"name": "<datastore name>"
# 		},
# 		"geo": "<geo>",
# 		"metadata": {
# 			"components_metadata": {
# 				"<column name1>": {
# 					"component_type": "column",
# 					"named_metadata": {
# 						"type": "length=10.0,nullable=true,type=date,scale=0.0,signed=false"
# 					}
# 				},
# 				"<column name2>": {
# 					"component_type": "column",
# 					"named_metadata": {
# 						"type": "length=3.0,nullable=true,type=char,scale=0.0,signed=false"
# 					},
# 					"tags": ["<tag1>", "<tag2>"]
# 				}
# 			},
# 			"dataset_named_metadata": {
# 				"<term1 name>": "<term1 value>",
# 				"<term2 name>": "<term2 value>"
# 			},
# 			"dataset_tags": [
# 				"<tag1>",
# 				"<tag2>"
# 			]
# 		},
# 	}
# }
Properties() = input.properties

Intent() = Properties().intent

Role() = Properties().role

AccessType() = input.type

DatasetTags() = input.details.metadata.dataset_tags

ProcessingGeo() = input.processing_geography

DestinationGeo() = input.destination



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