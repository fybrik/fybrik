package data_policies

#this file assumes input to be provided in specific format, in this case how data mesh provides it 
#similar file can be built for Egeria, at least for the metadata part, or any other catalog when we show how the input should be  parsed correctly

#TODO: add example of input struct

Purpose() = input.purpose 

Role() = input.role 

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