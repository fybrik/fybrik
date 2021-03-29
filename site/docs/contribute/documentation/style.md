# Style Guide

This page provides basic style guidance for keeping the documentation
of Mesh for Data **clear** and **understandable**. 

## Choose the right title

Use a short, keyword-rich title that captures the intent of the document and draws the reader in. Ensure that the title clearly and concisely conveys the content or subject matter and is meaningful to a global audience.

The text for the title of the document must use title case.
Capitalize the first letter of every word except conjunctions and prepositions.

|Do                      | Don't
|------------------------|-----
|`# Security Architecture` | `# Security architecture`
|`# Code of Conduct` | `# Code Of Conduct`

## Use sentence case for headings

Use sentence case for the headings in your document. Only capitalize the first
word of the heading, except for proper nouns or acronyms.

|Do                      | Don't
|------------------------|-----
|Configuring rate limits | Configuring Rate Limits
|Using Envoy for ingress | Using envoy for ingress
|Using HTTPS             | Using https


## Use present tense

|Do                           | Don't
|-----------------------------|------
|This command starts a proxy. | This command will start a proxy.

Exception: Use future or past tense if it is required to convey the correct
meaning. This exception is extremely rare and should be avoided.

## Use active voice

|Do                                         | Don't
|-------------------------------------------|------
|You can explore the API using a browser.   | The API can be explored using a browser.
|The YAML file specifies the replica count. | The replica count is specified in the YAML file.

## Use simple and direct language

Use simple and direct language. Avoid using unnecessary phrases, such as saying
"please."

|Do                          | Don't
|----------------------------|------
|To create a `ReplicaSet`, ... | In order to create a `ReplicaSet`, ...
|See the configuration file. | Please see the configuration file.
|View the Pods.              | With this next command, we'll view the Pods.

## Prefer shorter words over longer alternatives

|Do                                     | Don't
|---------------------------------------|------
|This tool helps scaling up pods.       | This tool facilitates scaling up pods.
|Pilot uses the `purpose` field to ...  | Pilot utilizes the `purpose` field to ... 

## Address the reader as "you"

|Do                                     | Don't
|---------------------------------------|------
|You can create a `Deployment` by ...     | We'll create a `Deployment` by ...
|In the preceding output, you can see...| In the preceding output, we can see ...

## Avoid using "we"

Using "we" in a sentence can be confusing, because the reader might not know
whether they're part of the "we" you're describing.

|Do                                        | Don't
|------------------------------------------|------
|Version 1.4 includes ...                  | In version 1.4, we have added ...
|Mesh for Data provides a new feature for ... | We provide a new feature ...
|This page teaches you how to use pods.    | In this page, we are going to learn about pods.

## Avoid jargon and idioms

Some readers speak English as a second language. Avoid jargon and idioms to help
make their understanding easier.

|Do                    | Don't
|----------------------|------
|Internally, ...       | Under the hood, ...
|Create a new cluster. | Turn up a new cluster.
|Initially, ...        | Out of the box, ...

## Avoid statements that will soon be out of date

Avoid using wording that becomes outdated quickly like "currently" and
"new". A feature that is new today is not new for long.

|Do                                  | Don't
|------------------------------------|------
|In version 1.4, ...                 | In the current version, ...
|The Federation feature provides ... | The new Federation feature provides ...

## Avoid statements about the future

Avoid making promises or giving hints about the future. If you need to talk about a feature in development, add a clear indication under the front matter that identifies the information accordingly:

!!! warning
    This page describes a feature that is not yet released

The only exceptions to this rule are design or architecture documents that can describe a vision. However, you must clearly distiquish between implemented features and a vision.

## Create useful links

There are good hyperlinks, and bad hyperlinks. The common practice of calling
links *here*  or *click here* are examples of bad hyperlinks. Check out [this
excellent article](https://medium.com/@heyoka/dont-use-click-here-f32f445d1021)
explaining what makes a good hyperlink and try to keep these guidelines in
mind when creating or reviewing site content.
