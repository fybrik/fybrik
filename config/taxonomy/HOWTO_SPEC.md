# Writing taxonomy specification

The project uses schema definitions to define taxonomy.
This document lists the requirements for writing such schemas.

## What spec to follow?

We use a subset of json schema DRAFT 4 that is described in this document. 

You SHOULD declare the schema version in the document:

```json
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    ...
}
```

The goal is to preserve compatibility with json schema DRAFT 4, OpenAPI 3.0, and Kubernetes CRDs (all are based on json schema DRAFT 4), and code generation tools.
We hope that newer json schema specs will be supported in the future as we improve our tooling.

## Schema properties

In this section we describe all schema properties that can be used. We categorize the properties to [structural properties](#structural-properties) and [validation properties](#validation-properties).

### Structural properties

The structural properties define the schema objects in terms of fields and types (cf. validation). They are the core of defining a
[_structural schema_](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema): a schema without polymorphism.

The following is a list of structural properties that can be used:

| Keyword | Type	| Description | Recommendation
| --- | --- | --- | ---
| type  | string | The type the schema: string, number, integer, boolean, array, or object. | MUST
| properties | object | Defines properties of the schema. Each value of this object MUST be a schema object. | MUST if defining a non-empty struct
| additionalProperties | bool or object | See [Field `additionalProperties`](#field-additionalproperties) | MUST if defining a map. SHOULD otherwise.
| items | object | Defines the schema of items in an array  | MUST if `type` is array
| description | string | Provide explanation about the purpose of the instance described by this schema. | SHOULD
| title | string | A short description | MAY
| default | value of `type` | Supply a default JSON value | MAY

The fields `properties`, `additionalProperties` with object type, and `items` are mutually exclusive:
1. Use `properties` to define object fields
    ```json
    // object with name and age
    {
        "type": "object",
        "properties": {
            "name": {
                "type": "string"
            },
            "age": {
                "type": "integer",
            }
        }
    }   
    ```
2. Use `additionalProperties` with object type to define a map
    ```json
    // map from string to integer values
    {
        "type": "object",
        "additionalProperties": {
            "type": "integer"
        }
    }
    ```
3. Use `items` to define the type of array items
    ```json
    // array of strings
    {
        "type": "array",
        "items": {
            "type": "string"
        }
    }
    ```
4. Omit all to define a primitive
   ```json
   {
    "type": "string",
   }
   ```

##### Field `additionalProperties`

The additionalProperties keyword is used to control the handling of extra stuff, that is, properties whose names are not listed in the properties keyword. 
The additionalProperties keyword may be either a boolean or an object:
1. If additionalProperties is a boolean and set to false, then no additional properties will be allowed. 
2. If additionalProperties is a boolean and set to true, then any additional properties are allowed.
3. If additionalProperties is an object then it defines a map between a `string` key to values of the type defined in `additionalProperties` as shown above. This is mutually exclusive with the use of `properties` and `items`.

Note: in json schema tooling the default is `additionalProperties: true` while in Kubernetes CRDs the default is equivalent to `additionalProperties: false`. We currently recommend setting this field explicitly. 

<!-- TODO: do we want `additionalProperties: true` as default always and change our tools accordingly? -->
<!-- TODO: update openapi2crd to drop `additionalProperties: false` and change `additionalProperties: true` to `x-kubernetes-preserve-unknown-fields: true` -->
<!-- TODO: implement a validation tool against the spec described here -->

### Validation properties

Validation properties augment the structural schema with properties that are used for validation against input documents.

This section describes properties for [combining schemas](#combining-schemas), the [`required`](#required-properties) property, and [value validation](#value-validation) properties.

#### Combining schemas

You can use the following keywords to create a complex schema, or validate a value against multiple criteria:

| Keyword | Type | Description |  
| ---     | ---  | --- |  
| allOf   | array | Validates the value against all the subschemas |
| oneOf   | array | Validates the value against exactly one of the subschemas |
| anyOf   | array | Validates the value against any (one or more) of the subschemas |
| not     | object | Makes sure the value is _not_ valid against the specified schema |

The subschemas within these keywords are intended for validation and not for changing the structural properties. 
For example, you can't define a new field, change type, or override the description. Specifically:
1. For each field in an object and each item in an array which is specified within any of `allOf`, `anyOf`, `oneOf` or `not`, the schema also specifies the field/item outside of those logical junctors.
2. The fields `description`, `type`, `default`, `additionalProperties` are not set in subschemas within `allOf`, `anyOf`, `oneOf` or `not`.

#### Required properties

By default, all properties defined in the `properties` field are optional and can be omitted. 
To mark a property as required add the property name to the `required` field.

| Keyword  | Type	| Description  
| ---      | ---     | ---         
| required | array  | List of required fields from `properties`

Examples:

```json
 // both lat and long are required properties
 {
    "type": "object",
    "properties": {
        "lat": {
            "type": "number"
        },
        "long": {
         "type": "number"
        }
    },
    "required": [
        "lat",
        "long"
    ]
 }
```

```json
// errors if foo and bar are both undefined
{
  "type": "object",
  "description": "foo bar object",
  "properties": {
    "foo": {
      "type": "string"
    },
    "bar": {
      "type": "integer"
    }
  },
  "anyOf": [
    {
      "required": [
        "foo"
      ]
    },
    {
      "required": [
        "bar"
      ]
    }
  ]
}
```

#### Value validation

| Keyword | Type	| Description 
| ---     | ---     | ---         
| format | string | See [Field `format`](#field-format) | MAY
| enum | array | An array of accepted values.
| multipleOf |   integer > 0 | A numeric instance is only valid if division by this keyword's value results in an integer.
| maximum | number | A numeric instance is only valid if is less than or exactly equal to the provided value
| ~~exclusiveMaximum~~ | bool | when set to true changes the value in `maximum` to be exclusive. NOT RECOMMENDED
| minimum | number |  A numeric instance is only valid if is greater than or exactly equal to the provided value
| ~~exclusiveMinimum~~ | bool | when set to true changes the value in `minimum` to be exclusive. NOT RECOMMENDED
| maxLength | integer >= 0 | A string instance is valid against this keyword if its length is less than, or equal to, the value of this keyword.
| minLength | integer >= 0 | A string instance is valid against this keyword if its length is greater than, or equal to, the value of this keyword.
| pattern | string | A string instance is considered valid if the provided [ECMA 262](https://www.ecma-international.org/ecma-262/5.1/#sec-7.8.5) regular expression matches the instance successfully.
| maxItems | integer >= 0 | An array instance is valid against "maxItems" if its size is less than, or equal to, the value of this keyword.
| minItems | integer >= 0 | An array instance is valid against "minItems" if its size is greater than, or equal to, the value of this keyword.
| uniqueItems | bool | When set to true an instance validates successfully if all of its elements are unique.
| maxProperties | integer >= 0 | An object instance is valid against "maxProperties" if its number of properties is less than, or equal to, the value of this keyword.
| minProperties | integer >= 0 | An object instance is valid against "minProperties" if its number of properties is greater than, or equal to, the value of this keyword.


##### Field `format`

Primitives have an optional modifier property: `format`. The format property is an open string-valued property and can have any value.

Below is a list of common formats that can be used in specific types. These MAY be enforced through validation though it's not guaranteed:

Common Name | [`type`](#dataTypes) | [`format`](#dataTypeFormat) | Comments
----------- | ------ | -------- | --------
integer | `integer` | `int32` | signed 32 bits
long | `integer` | `int64` | signed 64 bits
float | `number` | `float` | |
double | `number` | `double` | |
string | `string` | | |
byte | `string` | `byte` | base64 encoded characters
binary | `string` | `binary` | any sequence of octets
boolean | `boolean` | | |
date | `string` | `date` | As defined by `full-date` - [RFC3339](http://xml2rfc.ietf.org/public/rfc/html/rfc3339.html#anchor14)
dateTime | `string` | `date-time` | As defined by `date-time` - [RFC3339](http://xml2rfc.ietf.org/public/rfc/html/rfc3339.html#anchor14)
password | `string` | `password` | A hint to UIs to obscure input
email | `string` | `email` | Internet email address
hostname | `string` | `hostname` | an Internet host name
IPv4 | `string` | `ipv4` | IPv4 address according to the "dotted-quad" ABNF syntax
IPv6 | `string` | `ipv6` | IPv6 address 
URI  | `string` | `uri` | a valid URI, according to [RFC3986](https://tools.ietf.org/html/rfc3986)
UUID | `string` | `uuid` | a UUID

<!-- TODO: update openapi2crd to prune format if not one of: "int32", "int64", "float", "double", "byte", "date", "date-time", "password" -->

## Reference objects

Any time a schema object can be used, a reference object can be used in its place. This allows referencing definitions instead of defining them inline.

Assume that all schema files are mounted in a common folder so you can reference other taxonomy JSON files by their name. 

### Examples

A reference to a schema defined in an external file:

```json
{
  "$ref": "another.json#/definitions/Pet"
}
```

A reference to a schema defined in the same file:

```json
{
  "$ref": "#/definitions/Pet"
}
```

A more complete example:

```json
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "definitions": {
        "Part": {
            "type": "string",
            "enum": ["X", "Y"]
        },
        "System": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "part": {
                    "$ref": "#/definitions/Part"
                }
            }
        }
    }
}
```

## Additional reading material

This document is self contained, but you can read the following which are the source documentation that the rules in this document are augmented from:

- https://tools.ietf.org/html/draft-wright-json-schema-validation-00
- https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#schemaObject
- https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/
