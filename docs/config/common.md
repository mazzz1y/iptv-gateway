# Common Objects

This document describes common object types used across the configuration reference.

## Name/Value Object

| Field   | Type     | Required | Description                          |
|---------|----------|----------|--------------------------------------|
| `name`  | `string` | Yes      | Name identifier for the object       |
| `value` | `string` | Yes      | Value associated with the given name |

## Name/Patterns Object

| Field      | Type      | Required | Description                                  |
|------------|-----------|----------|----------------------------------------------|
| `name`     | `string`  | Yes      | Name identifier for the object               |
| `patterns` | `[]regex` | Yes      | Regular expression pattern(s) to match value |

## Name/Template Object

| Field      | Type         | Required | Description                    |
|------------|--------------|----------|--------------------------------|
| `name`     | `string`     | Yes      | Name identifier for the object |
| `template` | `gotemplate` | Yes      | Go template definition         |
