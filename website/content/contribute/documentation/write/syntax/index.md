---
title: Documentation Syntax
description: Documentation syntax is Markdown with a few extras.
date: 2020-04-26T22:00:53+03:00
draft: false
weight: 3
summary: Learn about the Markdown language used
---

The content is written in [Markdown language](https://www.markdownguide.org/) similar to `README.md` files in GitHub projects. However, there are a few extras you need to know about.

## Front Matter

The front matter is YAML code in between triple-dashed lines at the top of each file and provides important management options for the content. The following example shows a front matter with common fields filled by placeholders:

```plain
---
title: <title>
linktitle: <linktitle>
description: <description>
weight: <weight>
---
```

The available fields are described in the table below:

|Field              | Required   | Description                                                    |
|-------------------|------------|----------------------------------------------------------------|
|`title`            |   Yes      | The page's title.                                              |
|`linktitle`        |   No       | A shorter title that is used for links to the page.            |
|`description`      |   Yes      | A one-line description of the content on the page.             |
|`weight`           |   Yes      | The page order relative to the other pages in the directory.   |

## Add images

Place image files in the same directory as the markdown file using them. To make
localization easier and enhance accessibility, the preferred image
format is SVG. The following example shows the shortcode with the required
fields needed to add an image:

```plain
{{</* image width="75%" ratio="45.34%"
    src="./<image.svg>"
    caption="<The caption displayed under the image>"
    */>}}
```

The best way to create images and diagrams for this documentation is to use [draw.io](http://draw.io) as it 
allows users to later on edit the images/diagrams and is an open format. For this please use the web version or application
to edit the file and make sure to commit the .drawio file as well as the exported image file that is used in the documentation (e.g. png). 

The `src` and `caption` fields are required, but the shortcode also supports
optional fields, for example:

```plain
{{</* image width="75%" ratio="45.34%"
    src="./<image.svg>"
    link="<Link to open when the image is clicked>"
    alt="<Alternate text used by screen readers and when loading the image fails>"
    title="<Text that appears on mouse-over>"
    caption="<The caption displayed under the image>"
    */>}}
```

If you don't include the `link` field, Hugo uses the link set in `src`. 
If you don't include the `title` field, Hugo uses the text set in `caption`. If
you don't include the `alt` field, Hugo uses the text in `title` or in `caption`
if `title` is also not defined. 

The `width` field sets the size of the image relative to the surrounding text and
has a default of 100%.

The `ratio` field sets the height of the image relative to its width. Hugo
calculates this value automatically for image files in the folder.
However, you must calculate it manually for external images.
Set the value of `ratio` to `([image height]/[image width]) * 100`.


## Add links to other pages

The documentation supports three types of links depending on their target.
Each type uses a different syntax to express the target.

- **External links**. These are links to pages outside of {{< website >}}. Use the standard Markdown
  syntax to include the URL. Use the HTTPS protocol, when you reference files on the Internet, for example:

    ```plain
    [Descriptive text for the link](https://mysite/myfile.html)
    ```

- **Relative links**. These links target pages at the same level of the current
  file or further down the hierarchy. Start the path of relative links with a
  period `.`, for example:

    ```plain
    [This links to a sibling or child page](./sub-dir/child-page.html)
    ```

- **Absolute links**. These links target pages outside the hierarchy of the
  current page but within the {{< website >}} website. Start the path of absolute links
  with {{</* baseurl */>}}, for example:

    ```plain
    [This links to a page on the about section]({{</* baseurl */>}}/about/page/)
    ```

Regardless of type, links do not point to the `index.md` file with the content,
but to the folder containing it.

## Add links to content on GitHub

To refer to content in GitHub, use the following shortcodes: 

- `{{</* github_base */>}}` renders as the organization page `{{< github_base >}}`
- `{{</* github_repo */>}}` renders as the main repository `{{< github_repo >}}`

Tou can use these in external links. For example, to render a [README.md](https://{{< github_base >}}/{{< github_repo >}}/blob/master/README.md) link use:
  ```plain
  [README.md](https://{{</* github_base */>}}/{{</* github_repo */>}}/blob/master/README.md)
  ```

## Release information

To display current release information, use the following shortcodes: 

- `{{</* name */>}}` renders as {{< name >}}
- `{{</* version */>}}` renders as {{< version >}}
- `{{</* version_full */>}}` renders as {{< version_full >}}

## Callouts

To emphasize blocks of content, you can format them as warnings, tips, or
quotes. All callouts use very similar shortcodes:

```plain
{{</* warning */>}}
This is an important warning
{{</* /warning */>}}

{{</* tip */>}}
This is a useful tip from an expert
{{</* /tip */>}}

> This is a quote from somewhere
```

The shortcodes above render as follows:

{{< warning >}}
This is an important warning
{{< /warning >}}

{{< tip >}}
This is a useful tip from an expert
{{< /tip >}}

> This is a quote from somewhere

Use callouts sparingly. Each type of callout serves a specific purpose and
over-using them negates their intended purposes and their efficacy. Generally,
you should not include more than one callout per content file.

## Figures

The recommended format for figures is `SVG`, supported by many vector graphics tools like Google Draw, InkScape, and Illustrator. 

If your diagram depicts a process, do not add the descriptions of the steps to the diagram. Instead, only add the numbers of the steps to the diagram and add the descriptions of the steps as a numbered list in the document. Ensure that the numbers on the list match the numbers on your diagram. This approach helps make diagrams easier to understand and the content more accessible.

Number steps using circle graphics, using the bullet shortcode in the document. For example, the following produces  {{< bullet n=7 >}}:
```plain
{{</* bullet n=7 */>}}
```



