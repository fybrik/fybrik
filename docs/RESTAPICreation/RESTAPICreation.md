# Steps to creating a new REST API using Fybrik Taxonomy


## TL;DR

* Suppose you want to add a REST method  <span style="color:#FF0000">createNewComponent</span>  in the  <span style="color:#FF0000">datacatalog</span>  list of apis\.
* Edit <span style="color:red">connectors/api/datacatalog\.spec\.yaml</span>
* Edit <span style="color:red">pkg/model/datacatalog/api\.go</span>
* Run <span style="color:red">make generate</span> in fybrik root
  * Generates json files\. Remove <span style="color:red">external\.json\.</span> Keep the others.
* <span style="color:red">make generate\-client\-datacatalog</span>
  * Generates openapiclient files in <span style="color:red">pkg/connectors/datacatalog/openapiclient</span>
  * Add these two lines in <span style="color:red">pkg/connectors/datacatalog/openapiclient/model\.go</span>
    * <span style="color:#0000FF">type</span>  <span style="color:#000000"> </span>  <span style="color:#267F99">CreateNewComponentRequest</span>  <span style="color:#000000"> = </span>  <span style="color:#000000">datacatalog\.CreateNewComponentRequest</span>
    * <span style="color:#0000FF">type</span>  <span style="color:#000000"> </span>  <span style="color:#267F99">CreateNewComponentResponse</span>  <span style="color:#000000"> = </span>  <span style="color:#000000">datacatalog\.CreateNewComponentResponse</span>
* Define the new method in <span style="color:red">pkg/connectors/datacatalog/clients/datacatalog\.go</span>
  * <span style="color:#795E26">CreateNewComponent</span>  <span style="color:#000000">\(in \*</span>  <span style="color:#000000">datacatalog\.CreateNewComponentRequest</span>  <span style="color:#000000">\, creds </span>  <span style="color:#267F99">string</span>  <span style="color:#000000">\) \(\*</span>  <span style="color:#000000">datacatalog\.CreateNewComponentResponse</span>  <span style="color:#000000">\, </span>  <span style="color:#267F99">error</span>  <span style="color:#000000">\)</span>
* Add the implementation of <span style="color:#795E26"> </span>  <span style="color:#795E26">CreateNewComponent</span>  <span style="color:#795E26">\(\) method \(client side\) </span> in <span style="color:#795E26">datacatalog\_openapi\.go</span>
* Add server side implementation of  <span style="color:#795E26">CreateNewComponent</span>  in any connector
* <span style="color:#795E26">Add </span>  <span style="color:#795E26">CreateNewComponent</span>   implementation in <span style="color:#795E26">manager/controllers/</span>  <span style="color:#795E26">mockup</span>  <span style="color:#795E26"> </span>
* <span style="color:red">make verify </span> to fix tab issues </span>

## Longer Version


### <span style="color:blue">Step 1</span>: Define OpenAPI specification for the REST API

Suppose we want to create a new REST API:  <span style="color:#FF0000">createNewComponent</span>  under the  <span style="color:#FF0000">datacatalog</span>  group of APIs\.

Then we can define the API in  <span style="color:#FF0000">datacatalog\.spec\.yaml</span>  under  <span style="color:#FF0000">fybrik</span>  <span style="color:#FF0000">/connectors/</span>  <span style="color:#FF0000">api</span>  <span style="color:#FF0000"> </span> folder as follows:

![](img/Taxonomy-addRESTAPIinFybrik-v11.png)

### <span style="color:blue">Step 2</span>: Define request and response objects

Edit  <span style="color:#FF0000">pkg/model/</span>  <span style="color:#FF0000">datacatalog</span>  <span style="color:#FF0000">/</span>  <span style="color:#FF0000">api\.go</span>  <span style="color:#FF0000"> </span>

![](img/Taxonomy-addRESTAPIinFybrik-v12.png)

### <span style="color:blue">Step 3</span>: Generate the taxonomy files

Run  <span style="color:#FF0000">make generate </span> in  <span style="color:#FF0000">fybrik</span>  <span style="color:#FF0000"> root </span> folder

  * Generates json files\. Remove external\.json\. Keep the others

### <span style="color:blue">Step 4</span>: Generate OpenApiClient related client code

  * Run the following command in  <span style="color:#FF0000">connectors/api</span>  folder to generate openapiclient files in  <span style="color:#FF0000">pkg/connectors/datacatalog/openapiclient</span>
    * <span style="color:red">	make generate\-client\-datacatalog</span>
  * Add these two lines in  <span style="color:#FF0000">pkg/connectors/datacatalog/openapiclient/model\.go</span>
    * <span style="color:#0000FF">	type</span>  <span style="color:#000000"> </span>  <span style="color:#267F99">CreateNewComponentRequest</span>  <span style="color:#000000"> = </span>  <span style="color:#000000">datacatalog\.CreateNewComponentRequest</span>
    * <span style="color:#0000FF">	type</span>  <span style="color:#000000"> </span>  <span style="color:#267F99">CreateNewComponentResponse</span>  <span style="color:#000000"> = </span>  <span style="color:#000000">datacatalog\.CreateNewComponentResponse</span>

### <span style="color:blue">Step 5</span>: Define and Implement the new REST API in the datacatalog interface

Define the new method  <span style="color:#FF0000">CreateComponent</span>  in  <span style="color:#FF0000">datacatalog\.go</span>  <span style="color:#FF0000"> </span> in  <span style="color:#FF0000">pkg/connectors/datacatalog/clients/</span>

![](img/Taxonomy-addRESTAPIinFybrik-v13.png)

Add the implementation of <span style="color:#795E26"> </span>  <span style="color:#795E26">CreateNewComponent</span>  <span style="color:#795E26">\(\) method \(client side\) in </span>  <span style="color:#795E26">datacatalog\_openapi\.go</span>  <span style="color:#795E26"> </span>

![](img/Taxonomy-addRESTAPIinFybrik-v14.png)

### <span style="color:blue">Step 6</span>: Add server side  logic for the REST API

Add server side implementation of  <span style="color:#795E26">CreateNewComponent</span>  <span style="color:#795E26">\(\) in any connector </span>

![](img/Taxonomy-addRESTAPIinFybrik-v15.png)

### <span style="color:blue">Step 7</span>: Add test / dummy implementation of REST API in manager/mockup

<span style="color:#795E26">Add </span>  <span style="color:#795E26">CreateNewComponent</span>  <span style="color:#795E26"> implementation in manager/controllers/</span>  <span style="color:#795E26">mockup</span>  <span style="color:#795E26"> </span>

![](img/Taxonomy-addRESTAPIinFybrik-v16.png)

### <span style="color:blue">Step 8</span>: Verify the changes

<span style="color:#3B2322">Run </span>  <span style="color:#FF0000">make verify </span>  <span style="color:#3B2322">in </span>  <span style="color:#3B2322">fybrik</span>  <span style="color:#3B2322"> root folder to fix go\-linting / go\-compilation issues </span>


### Discussions

The example used in this document is implemented in this github branch for reference\. Please check this link for more details : [https://github\.com/rohithdv/fybrik/tree/taxonomy\-kt](https://github.com/rohithdv/fybrik/tree/taxonomy-kt)

