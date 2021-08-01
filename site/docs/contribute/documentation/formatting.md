# Formatting Standards

This page shows the formatting standards for the Fybrik documentation.

## Link to other pages using relative links

When linking between pages in the documentation you can simply use the regular Markdown linking syntax, including the relative path to the Markdown document you wish to link to. For example:

```markdown
Please see the [project license](license.md) for further details.
```

If the target documentation file is in another directory you'll need to make sure to include any relative directory path in the link:

```markdown
Please see the [project license](../about/license.md) for further details.
```

## Prefer SVG format for diagrams

Place image files in the [docs/static](https://github.com/fybrik/fybrik/tree/master/site/docs/static) directory. Use regular Markdown syntax for images. For example:

```markdown
![](../static/myimage.svg)
```

To make localization easier and enhance accessibility, the preferred image format is SVG. We recommend to use [draw.io](https://draw.io) for creating images and diagrams. Use **Export as** to save your image in SVG format. Keep the **Include a copy of my diagram** option checked to allow later loading the SVG in draw.io and be sure to check **Embed images** if you diagram includes any.

If your diagram depicts a process, try to avoid adding the descriptions of the steps to the diagram. Instead, only add the numbers of the steps to the diagram and add the descriptions of the steps as a numbered list in the document. Ensure that the numbers on the list match the numbers on your diagram. This approach helps make diagrams easier to understand and the content more accessible.

## Do not wrap lines

Never wrap lines after a fixed number of characters or in a middle of a senstence.
Instead, configure your editor to soft-wrap when editing documentation.

|Do                | Don't
|------------------|------
| This is a long line.   | This is a <br>long line.


## Use angle brackets for placeholders

Use angle brackets for placeholders in commands or code samples. Tell the reader
what the placeholder represents. For example:


1. Display information about a pod:
    ```bash
    $ kubectl describe pod <pod-name>
    ```
    Where `<pod-name>` is the name of one of your pods.

## Use **bold** to emphasize user interface elements

|Do                | Don't
|------------------|------
|Click **Fork**.   | Click "Fork".
|Select **Other**. | Select 'Other'.

## Use **bold** to emphasize important text

Use **bold** to emphasize text that is particularly important. Avoid overusing bold as it reduces its impact and readability. 

| Do | Don't | 
| - | - |
|  Examples of **bad** configurations: | Examples of **bad configurations**: |
|  The maximum length of the `name` field is **256 characters**. | **The maximum length of the `name` field is 256 characters**.  |

## Don't use capitalization for emphasis

Only use the original capitalization found in the code or configuration files
when referencing those values directly. Use back-ticks \`\` around the
referenced value to make the connection explicit. For example, use
`IsolationPolicy`, not `Isolation Policy` or `isolation policy`.

If you are not referencing values or code directly, use normal sentence
capitalization, for example, "The isolation policy configuration takes place
in a YAML file."

## Use _italics_ to emphasize new terms

|Do                                         | Don't
|-------------------------------------------|---
|A _cluster_ is a set of nodes ...          | A "cluster" is a set of nodes ...
|These components form the _control plane_. | These components form the **control plane**.

## Use `back-ticks` around file names, directories, and paths

|Do                                   | Don't
|-------------------------------------|------
|Open the `foo.yaml` file.         | Open the foo.yaml file.
|Go to the `/content/docs/architecture` directory.  | Go to the /content/docs/architecture directory.
|Open the `/data/args.yaml` file. | Open the /data/args.yaml file.

## Use `back-ticks` around inline code and commands

|Do                          | Don't
|----------------------------|------
|The `foo run` command creates a `Deployment`. | The "foo run" command creates a `Deployment`.
|For declarative management, use `foo apply`. | For declarative management, use "foo apply".

Use code-blocks for commands you intend readers to execute. Only use inline code
and commands to mention specific labels, flags, values, functions, objects,
variables, modules, or commands.

## Use `back-ticks` around object field names

|Do                                                               | Don't
|-----------------------------------------------------------------|------
|Set the value of the `ports` field in the configuration file. | Set the value of the "ports" field in the configuration file.
|The value of the `rule` field is a `Rule` object.           | The value of the "rule" field is a `Rule` object.
